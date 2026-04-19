package treesitter

import (
	"path"
	"strings"
	"unsafe"

	ts "github.com/tree-sitter/go-tree-sitter"
	tree_sitter_c "github.com/tree-sitter/tree-sitter-c/bindings/go"
	tree_sitter_cpp "github.com/tree-sitter/tree-sitter-cpp/bindings/go"
	tree_sitter_csharp "github.com/tree-sitter/tree-sitter-c-sharp/bindings/go"
	tree_sitter_go "github.com/tree-sitter/tree-sitter-go/bindings/go"
	tree_sitter_java "github.com/tree-sitter/tree-sitter-java/bindings/go"
	tree_sitter_javascript "github.com/tree-sitter/tree-sitter-javascript/bindings/go"
	tree_sitter_php "github.com/tree-sitter/tree-sitter-php/bindings/go"
	tree_sitter_python "github.com/tree-sitter/tree-sitter-python/bindings/go"
	tree_sitter_ruby "github.com/tree-sitter/tree-sitter-ruby/bindings/go"
	tree_sitter_rust "github.com/tree-sitter/tree-sitter-rust/bindings/go"
	tree_sitter_typescript "github.com/tree-sitter/tree-sitter-typescript/bindings/go"
)

type langDef struct {
	language *ts.Language
	importQ  string
	defQ     string
}

var languages map[string]*langDef

func init() {
	languages = make(map[string]*langDef)

	goLang := makeLang(tree_sitter_go.Language(),
		`(import_spec path: (interpreted_string_literal) @import)`,
		`[
			(function_declaration name: (identifier) @def)
			(method_declaration name: (field_identifier) @def)
			(type_spec name: (type_identifier) @def)
		]`,
	)
	register(goLang, ".go")

	jsLang := makeLang(tree_sitter_javascript.Language(),
		`[
			(import_statement source: (string) @import)
			(call_expression function: (identifier) @_fn (#eq? @_fn "require") arguments: (arguments (string) @import))
		]`,
		`[
			(function_declaration name: (identifier) @def)
			(class_declaration name: (identifier) @def)
			(variable_declarator name: (identifier) @def)
		]`,
	)
	register(jsLang, ".js", ".jsx", ".mjs", ".cjs")

	tsLang := makeLang(tree_sitter_typescript.LanguageTypescript(),
		`[
			(import_statement source: (string) @import)
			(call_expression function: (identifier) @_fn (#eq? @_fn "require") arguments: (arguments (string) @import))
		]`,
		`[
			(function_declaration name: (identifier) @def)
			(class_declaration name: (identifier) @def)
			(interface_declaration name: (type_identifier) @def)
			(type_alias_declaration name: (type_identifier) @def)
			(variable_declarator name: (identifier) @def)
		]`,
	)
	register(tsLang, ".ts")

	tsxLang := makeLang(tree_sitter_typescript.LanguageTSX(),
		tsLang.importQ,
		tsLang.defQ,
	)
	register(tsxLang, ".tsx")

	pyLang := makeLang(tree_sitter_python.Language(),
		`[
			(import_statement name: (dotted_name) @import)
			(import_from_statement module_name: (dotted_name) @import)
		]`,
		`[
			(function_definition name: (identifier) @def)
			(class_definition name: (identifier) @def)
		]`,
	)
	register(pyLang, ".py")

	javaLang := makeLang(tree_sitter_java.Language(),
		`(import_declaration (scoped_identifier) @import)`,
		`[
			(method_declaration name: (identifier) @def)
			(class_declaration name: (identifier) @def)
			(interface_declaration name: (identifier) @def)
		]`,
	)
	register(javaLang, ".java")

	csLang := makeLang(tree_sitter_csharp.Language(),
		`(using_directive (qualified_name) @import)`,
		`[
			(method_declaration name: (identifier) @def)
			(class_declaration name: (identifier) @def)
			(interface_declaration name: (identifier) @def)
		]`,
	)
	register(csLang, ".cs")

	rustLang := makeLang(tree_sitter_rust.Language(),
		`(use_declaration argument: (_) @import)`,
		`[
			(function_item name: (identifier) @def)
			(struct_item name: (type_identifier) @def)
			(enum_item name: (type_identifier) @def)
			(trait_item name: (type_identifier) @def)
			(impl_item type: (type_identifier) @def)
		]`,
	)
	register(rustLang, ".rs")

	phpLang := makeLang(tree_sitter_php.LanguagePHP(),
		`(use_declaration (qualified_name) @import)`,
		`[
			(function_definition name: (name) @def)
			(class_declaration name: (name) @def)
			(interface_declaration name: (name) @def)
		]`,
	)
	register(phpLang, ".php")

	rubyLang := makeLang(tree_sitter_ruby.Language(),
		`(call method: (identifier) @_fn (#match? @_fn "require") arguments: (argument_list (string) @import))`,
		`[
			(method name: (identifier) @def)
			(class name: (constant) @def)
			(module name: (constant) @def)
		]`,
	)
	register(rubyLang, ".rb")

	cLang := makeLang(tree_sitter_c.Language(),
		`(preproc_include path: (_) @import)`,
		`(function_definition declarator: (function_declarator declarator: (identifier) @def))`,
	)
	register(cLang, ".c", ".h")

	cppLang := makeLang(tree_sitter_cpp.Language(),
		`[
			(preproc_include path: (_) @import)
			(using_declaration (qualified_identifier) @import)
		]`,
		`[
			(function_definition declarator: (function_declarator declarator: (identifier) @def))
			(class_specifier name: (type_identifier) @def)
		]`,
	)
	register(cppLang, ".cpp", ".cc", ".cxx", ".hpp", ".hh")
}

func makeLang(ptr unsafe.Pointer, importQ, defQ string) *langDef {
	return &langDef{
		language: ts.NewLanguage(ptr),
		importQ:  importQ,
		defQ:     defQ,
	}
}

func register(ld *langDef, exts ...string) {
	for _, ext := range exts {
		languages[ext] = ld
	}
}

// LangForFile returns the language definition for a filename, or nil if unsupported.
func LangForFile(filename string) *langDef {
	ext := strings.ToLower(path.Ext(filename))
	return languages[ext]
}
