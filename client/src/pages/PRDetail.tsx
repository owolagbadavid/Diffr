import { useCallback, useEffect, useRef, useState } from "react";
import { useNavigate, useParams } from "react-router-dom";
import type { PRFilesResponse, Strategy } from "../types";
import { fetchPRFiles, fetchStrategies, FetchError } from "../api";
import FileCard from "../components/FileCard";
import FileViewer from "../components/FileViewer";
import SymbolPopover from "../components/SymbolPopover";
import SearchPanel from "../components/SearchPanel";

export default function PRDetail() {
  const { owner, repo, number } = useParams<{
    owner: string;
    repo: string;
    number: string;
  }>();
  const navigate = useNavigate();
  const prNumber = Number(number);

  const [data, setData] = useState<PRFilesResponse | null>(null);
  const [strategies, setStrategies] = useState<Strategy[]>([]);
  const [strategy, setStrategy] = useState("default");
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState("");
  const [notFound, setNotFound] = useState(false);

  const [viewerFile, setViewerFile] = useState<{
    url: string;
    name: string;
  } | null>(null);

  // Symbol search state
  const [popover, setPopover] = useState<{
    word: string;
    rect: DOMRect;
  } | null>(null);
  const [highlightWord, setHighlightWord] = useState<string | null>(null);
  const [searchQuery, setSearchQuery] = useState<string | null>(null);
  const [prSearchResults, setPRSearchResults] = useState<
    { filename: string; line: string; lineNum: number }[] | null
  >(null);

  const fileCache = useRef<Record<string, string[]>>({});

  useEffect(() => {
    fetchStrategies()
      .then(setStrategies)
      .catch(() => {});
  }, []);

  useEffect(() => {
    if (!owner || !repo || !prNumber) return;
    setLoading(true);
    setError("");
    setNotFound(false);
    fetchPRFiles(owner, repo, prNumber, strategy)
      .then(setData)
      .catch((err) => {
        if (err instanceof FetchError && err.message.includes("404")) {
          setNotFound(true);
        } else if (err instanceof FetchError) {
          setError(err.message);
        } else {
          setError("Failed to load PR files");
        }
      })
      .finally(() => setLoading(false));
  }, [owner, repo, prNumber, strategy]);

  const handleTokenClick = useCallback(
    (word: string, rect: DOMRect) => {
      setPopover({ word, rect });
    },
    []
  );

  function handleHighlightAll() {
    if (popover) {
      setHighlightWord(popover.word);
      setPopover(null);
    }
  }

  function handleFindInPR() {
    if (!popover || !data) return;
    const word = popover.word;
    const results: { filename: string; line: string; lineNum: number }[] = [];

    for (const group of data.groups) {
      for (const file of group.files) {
        if (!file.patch) continue;
        const lines = file.patch.split("\n");
        let newLine = 0;
        for (const line of lines) {
          const hunkMatch = line.match(
            /^@@ -\d+(?:,\d+)? \+(\d+)(?:,\d+)? @@/
          );
          if (hunkMatch) {
            newLine = parseInt(hunkMatch[1], 10);
            continue;
          }
          if (line.startsWith("+") || line.startsWith(" ")) {
            if (line.includes(word)) {
              results.push({
                filename: file.filename,
                line: line.slice(1),
                lineNum: newLine,
              });
            }
            newLine++;
          } else if (line.startsWith("-")) {
            // deleted lines don't increment newLine
          }
        }
      }
    }

    setPRSearchResults(results);
    setHighlightWord(word);
    setPopover(null);
  }

  function handleSearchRepo() {
    if (popover) {
      setSearchQuery(popover.word);
      setPopover(null);
    }
  }

  if (notFound) {
    return (
      <div className="text-center mt-16">
        <h2 className="text-xl font-semibold mb-2">Pull request not found</h2>
        <p className="text-(--muted) mb-4">
          PR #{number} doesn't exist in{" "}
          <span className="font-mono text-(--accent)">
            {owner}/{repo}
          </span>
          .
        </p>
        <button
          onClick={() => navigate(`/${owner}/${repo}`)}
          className="px-4 py-2 bg-(--accent) text-white rounded-lg text-sm font-medium hover:opacity-85 transition-opacity"
        >
          Back to PRs
        </button>
      </div>
    );
  }

  return (
    <div>
      <div className="flex items-center justify-between mb-4">
        <h2 className="text-lg font-semibold">
          #{number} — {data ? `${data.total_files} files` : ""}
        </h2>
        {highlightWord && (
          <button
            onClick={() => {
              setHighlightWord(null);
              setPRSearchResults(null);
            }}
            className="text-xs text-(--muted) hover:text-(--text) px-2 py-1 border border-(--border) rounded-lg transition-colors"
          >
            Clear highlight: <span className="font-mono text-(--accent)">{highlightWord}</span>
          </button>
        )}
      </div>

      <div className="flex items-center gap-2 mb-5">
        <label className="text-sm text-(--muted)">Strategy:</label>
        <select
          value={strategy}
          onChange={(e) => setStrategy(e.target.value)}
          className="px-3 py-2 bg-(--surface) border border-(--border) rounded-lg text-(--text) text-sm outline-none cursor-pointer focus:border-(--accent)"
        >
          {strategies.length > 0 ? (
            strategies.map((s) => (
              <option key={s.name} value={s.name}>
                {s.name} — {s.description}
              </option>
            ))
          ) : (
            <option value="default">default</option>
          )}
        </select>
      </div>

      {/* PR search results (find in PR files) */}
      {prSearchResults && prSearchResults.length > 0 && (
        <div className="mb-5 border border-(--border) rounded-lg overflow-hidden">
          <div className="flex items-center justify-between px-3 py-2 bg-(--surface) text-[13px]">
            <span className="text-(--muted)">
              Found <span className="text-(--text) font-medium">{prSearchResults.length}</span> matches in PR for{" "}
              <span className="font-mono text-(--accent)">{highlightWord}</span>
            </span>
            <button
              onClick={() => setPRSearchResults(null)}
              className="text-(--accent) hover:underline bg-transparent text-[13px]"
            >
              Dismiss
            </button>
          </div>
          <div className="max-h-48 overflow-auto">
            {prSearchResults.map((r, i) => (
              <div
                key={i}
                className="flex items-baseline gap-3 px-3 py-1.5 text-xs font-mono border-t border-(--border) hover:bg-(--surface)"
              >
                <span className="text-(--accent) shrink-0 truncate max-w-[250px]">
                  {r.filename}
                </span>
                <span className="text-(--muted) shrink-0">:{r.lineNum}</span>
                <span className="text-(--text) truncate">{r.line.trim()}</span>
              </div>
            ))}
          </div>
        </div>
      )}

      {loading && (
        <div className="text-center text-(--muted) py-8">Loading...</div>
      )}

      {error && (
        <div className="bg-(--red)/10 border border-(--red)/30 text-(--red) px-4 py-3 rounded-lg text-sm mt-4">
          {error}
        </div>
      )}

      {!loading &&
        !error &&
        data?.groups.map((group) => (
          <div key={group.name} className="mb-6">
            <div className="text-[13px] font-semibold text-(--muted) uppercase tracking-wider pb-2 border-b border-(--border) mb-2">
              {group.name} ({group.files.length})
            </div>
            {group.files.map((f) => (
              <FileCard
                key={f.filename}
                file={f}
                fileCache={fileCache}
                highlightWord={highlightWord}
                onViewFile={(url, name) => setViewerFile({ url, name })}
                onTokenClick={handleTokenClick}
              />
            ))}
          </div>
        ))}

      {popover && (
        <SymbolPopover
          word={popover.word}
          rect={popover.rect}
          onHighlightAll={handleHighlightAll}
          onFindInPR={handleFindInPR}
          onSearchRepo={handleSearchRepo}
          onClose={() => setPopover(null)}
        />
      )}

      {viewerFile && (
        <FileViewer
          contentsURL={viewerFile.url}
          filename={viewerFile.name}
          fileCache={fileCache}
          highlightWord={highlightWord}
          onTokenClick={handleTokenClick}
          onClose={() => setViewerFile(null)}
        />
      )}

      {searchQuery && owner && repo && (
        <SearchPanel
          owner={owner}
          repo={repo}
          query={searchQuery}
          onClose={() => setSearchQuery(null)}
        />
      )}
    </div>
  );
}
