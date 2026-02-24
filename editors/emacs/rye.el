;;; rye.el --- Emacs major mode for the Rye programming language -*- lexical-binding: t; -*-

;;; Commentary:
;;
;; Major mode for editing Rye programming language files.
;; https://ryelang.org / https://github.com/refaktor/rye
;;
;; Features:
;;   - Syntax highlighting for all Rye word types
;;   - Smart indentation based on block nesting
;;   - Comment support (;)
;;   - Bracket matching for {}, [], ()
;;
;; Installation:
;;   (load "/path/to/rye.el")
;;   or add to load-path and (require 'rye)
;;
;; History:
;;   2021 - Janko Metelko <janko.itm@gmail.com> - initial Rye adaptation from Rebol mode
;;   2025 - Rewritten for comprehensive Rye language support
;;

;;; Code:

(defgroup rye nil
  "Major mode for the Rye programming language."
  :group 'languages
  :prefix "rye-")

(defcustom rye-indent-offset 4
  "Number of spaces for each indentation level in Rye."
  :type 'integer
  :group 'rye)

(defcustom rye-command "rye"
  "Shell command to run the Rye interpreter."
  :type 'string
  :group 'rye)

;; ============================================================
;; Syntax table
;; ============================================================

(defvar rye-mode-syntax-table
  (let ((st (make-syntax-table)))
    ;; Comments: ; to end of line
    (modify-syntax-entry ?\; "<" st)
    (modify-syntax-entry ?\n ">" st)

    ;; Strings
    (modify-syntax-entry ?\" "\"" st)
    (modify-syntax-entry ?` "\"" st)

    ;; Brackets
    (modify-syntax-entry ?\{ "(}" st)
    (modify-syntax-entry ?\} "){" st)
    (modify-syntax-entry ?\[ "(]" st)
    (modify-syntax-entry ?\] ")[" st)
    (modify-syntax-entry ?\( "()" st)
    (modify-syntax-entry ?\) ")(" st)

    ;; These are word constituents in Rye
    (modify-syntax-entry ?- "w" st)
    (modify-syntax-entry ?? "w" st)
    (modify-syntax-entry ?! "w" st)
    (modify-syntax-entry ?* "w" st)
    (modify-syntax-entry ?+ "w" st)
    (modify-syntax-entry ?= "w" st)
    (modify-syntax-entry ?< "w" st)
    (modify-syntax-entry ?> "w" st)
    (modify-syntax-entry ?~ "w" st)
    (modify-syntax-entry ?^ "w" st)
    (modify-syntax-entry ?\\ "w" st)

    ;; Punctuation
    (modify-syntax-entry ?| "." st)
    (modify-syntax-entry ?. "." st)
    (modify-syntax-entry ?, "." st)
    (modify-syntax-entry ?/ "." st)
    (modify-syntax-entry ?: "." st)
    (modify-syntax-entry ?@ "." st)
    (modify-syntax-entry ?% "." st)
    (modify-syntax-entry ?' "." st)

    st)
  "Syntax table for `rye-mode'.")

;; ============================================================
;; Font-lock (syntax highlighting)
;; ============================================================

;; --- Keyword lists ---

(defconst rye--keywords-function-def
  '("fn" "fn1" "fnc" "does" "pfn" "closure"
    "fn\\cc" "fn\\in" "fn\\inside" "partial" "method")
  "Function definition keywords.")

(defconst rye--keywords-control-flow
  '("if" "either" "switch" "cases" "choose" "when" "return"
    "forever" "forever\\with")
  "Control flow keywords.")

(defconst rye--keywords-iteration
  '("for" "for\\pos" "for\\idx" "for\\" "for\\kv"
    "loop" "while" "until"
    "walk" "walk\\pos" "walk\\idx"
    "produce" "produce\\while" "produce\\"
    "replicate" "replicate\\idx"
    "recur-if" "recur-if\\1" "recur-if\\2" "recur-if\\3")
  "Iteration keywords.")

(defconst rye--keywords-error-handling
  '("fail" "failure" "failure\\wrap"
    "fix" "fix\\either" "fix\\else" "fix\\continue" "fix\\match"
    "check" "ensure" "requires-one-of"
    "disarm" "try" "try-all" "try\\in"
    "finally" "retry" "persist" "timeout" "continue"
    "is-error" "is-failure" "is-success" "has-failed"
    "error-kind?" "cause?" "status?" "message?" "details?")
  "Error handling keywords.")

(defconst rye--keywords-context
  '("context" "context\\pure" "raw-context" "isolate"
    "private" "private\\"
    "extends" "bind!" "unbind" "clone" "clone\\" "clone\\deep"
    "do" "do\\in" "do\\inside" "do\\inx"
    "with" "current" "parent?" "parent\\of"
    "cc" "ccp" "ccb" "mkcc"
    "lc" "lcp" "lc\\" "lcp\\" "lc\\data" "lc\\data\\"
    "whereis" "get" "import" "load" "load\\mod" "load\\live" "load\\sig"
    "use" "rye")
  "Context-related keywords.")

(defconst rye--keywords-combinators
  '("pass" "keep" "wrap" "apply" "evals" "evals\\with" "on-change")
  "Combinator keywords.")

(defconst rye--keywords-types
  '("dict" "list" "table" "vector" "ref" "deref" "is-ref"
    "secret" "reveal" "kind" "assure-kind" "complex" "lazy")
  "Type and constructor keywords.")

(defconst rye--keywords-other
  '("var" "set" "unset!" "val" "change!" "modify!"
    "defer" "defer\\"
    "assert" "assert\\display"
    "doc!" "doc?" "doc\\of?"
    "save\\current" "save\\current\\secure"
    "scmd" "scmd\\capture")
  "Other keywords.")

(defconst rye--builtins-collections
  '("map" "map\\pos" "map\\idx"
    "filter" "purge" "purge!" "seek"
    "reduce" "fold" "fold\\do"
    "partition" "group"
    "sort" "sort!" "sort\\by" "sort\\by\\key"
    "unique" "reverse" "reverse!"
    "union" "intersection" "intersection\\by" "difference"
    "zip" "transpose" "unpack" "sample" "random"
    "first" "second" "third" "last"
    "rest" "rest\\from" "tail" "head" "before-last" "nth"
    "range" "concat"
    "append!" "append\\many!"
    "update!" "update\\with!" "update\\pos!"
    "remove-last!" "change\\nth!"
    "keys" "values" "peek" "pop" "pos" "next" "at"
    "is-empty" "length?" "max-idx?")
  "Collection builtins.")

(defconst rye--builtins-strings
  '("trim" "trim\\" "trim\\right" "trim\\left"
    "replace" "substring" "contains" "has-suffix" "has-prefix"
    "index?" "position?"
    "split" "split\\quoted" "split\\many" "split\\every"
    "join" "join\\with"
    "capitalize" "lower" "upper"
    "concat3" "space" "ln"
    "encode-to\\base64" "decode\\base64"
    "regexp" "match" "submatch" "find-all")
  "String builtins.")

(defconst rye--builtins-printing
  '("print" "print2" "prn" "prn2" "prns" "prns2"
    "prnf" "printf" "prnv" "printv"
    "print\\ssv" "print\\csv"
    "probe" "probe\\" "inspect"
    "format" "embed"
    "display" "display\\custom" "display-selection" "display-date-input"
    "esc" "esc-val" "capture-stdout"
    "mold" "mold\\nowrap" "dump"
    "newline" "tab" "input")
  "Printing builtins.")

(defconst rye--builtins-conversion
  '("to-integer" "to-decimal" "to-string" "to-uri" "to-file"
    "to-char" "to-block" "to-context" "to-word" "to-json" "to-table"
    "type?" "types?" "kind?"
    "is-string" "is-integer" "is-decimal" "is-number"
    "is-positive" "is-zero" "is-even" "is-odd"
    "is-multiple-of" "is-error-of-kind"
    "parse-json" "autotype")
  "Conversion builtins.")

(defconst rye--builtins-math
  '("inc" "dec" "inc!" "dec!"
    "negate" "invert" "sign" "mod" "abs"
    "sum" "mul" "avg" "max" "min"
    "max\\by" "min\\by" "avg\\by" "sum\\by"
    "clamp" "add" "addnums"
    "random\\integer" "random\\decimal"
    "normalize" "std-deviation"
    "cosine-similarity" "correlation"
    "dot-product" "euclidean-distance"
    "mean-vectors" "unit-vector" "project-vector" "reject-vector")
  "Math builtins.")

(defconst rye--builtins-time
  '("now" "date" "datetime" "sleep" "time-it"
    "format-date" "format-imap-date"
    "year?" "month?" "day?" "hour?" "minute?" "second?"
    "weekday?" "yearday?" "date?" "time?" "days-in-month?"
    "unix-micro?" "unix-milli?"
    "seconds" "minutes" "hours" "days" "weeks"
    "thousands" "millions")
  "Time builtins.")

(defconst rye--builtins-io
  '("Read" "Write" "Write*" "Open" "Open\\append" "Close"
    "Does-exist" "File-ext?" "Read\\lines" "Read\\string"
    "ls" "cd" "cwd?" "mkdir")
  "IO builtins.")

(defconst rye--builtins-testing
  '("section" "group" "equal" "error" "stdout")
  "Testing builtins.")

(defconst rye--builtins-other
  '("collect" "collect!" "collected?" "collector" "needs"
    "eyr" "eyr\\full" "eyr\\loop" "eyr\\clear" "calc"
    "match" "match-block" "validate" "validate>ctx"
    "cmd" "Run" "Request" "Call" "Header!" "Serve" "Handle"
    "Reader" "Get"
    "markdown->html" "do-sxml" "do-markdown" "Attr?"
    "Load\\csv" "columns?" "order-by"
    "where-contains" "where-equal" "where-match"
    "where-greater" "where-lesser" "where-between")
  "Other builtins.")

(defconst rye--constants
  '("true" "false" "none" "nil")
  "Language constants.")

(defconst rye--logical-operators
  '("and" "or" "xor" "not" "all" "any")
  "Logical operator words.")

;; --- Helper to build regexp from word lists ---

(defun rye--words-regexp (words)
  "Build a regexp matching any of WORDS as whole Rye words."
  (concat "\\(?:^\\|[[:space:]{}()\\[,]\\)"
          "\\(" (regexp-opt words t) "\\)"
          "\\(?:[[:space:]{}()\\],;]\\|$\\)"))

;; --- Font-lock keywords ---

(defvar rye-font-lock-keywords
  (let ((kw-fn-def    (regexp-opt rye--keywords-function-def 'words))
        (kw-control   (regexp-opt rye--keywords-control-flow 'words))
        (kw-iter      (regexp-opt rye--keywords-iteration 'words))
        (kw-error     (regexp-opt rye--keywords-error-handling 'words))
        (kw-ctx       (regexp-opt rye--keywords-context 'words))
        (kw-comb      (regexp-opt rye--keywords-combinators 'words))
        (kw-types     (regexp-opt rye--keywords-types 'words))
        (kw-other     (regexp-opt rye--keywords-other 'words))
        (bi-coll      (regexp-opt rye--builtins-collections 'words))
        (bi-str       (regexp-opt rye--builtins-strings 'words))
        (bi-print     (regexp-opt rye--builtins-printing 'words))
        (bi-conv      (regexp-opt rye--builtins-conversion 'words))
        (bi-math      (regexp-opt rye--builtins-math 'words))
        (bi-time      (regexp-opt rye--builtins-time 'words))
        (bi-io        (regexp-opt rye--builtins-io 'words))
        (bi-test      (regexp-opt rye--builtins-testing 'words))
        (bi-other     (regexp-opt rye--builtins-other 'words))
        (constants    (regexp-opt rye--constants 'words))
        (logical-ops  (regexp-opt rye--logical-operators 'words)))
    (list

     ;; --- Comments (handled by syntax table, but shebang needs explicit match) ---
     '("\\`#!.*$" . font-lock-comment-face)

     ;; --- Strings: URIs ---
     '("\\(?:^\\|[[:space:]{(\\[]\\)\\([a-zA-Z][a-zA-Z0-9+.-]*://[^[:space:]{}\\[\\]]*\\)" 1 font-lock-string-face)

     ;; --- File paths: %word ---
     '("\\(?:^\\|[[:space:]{(\\[]\\)\\(%[^[:space:]{}\\[\\]]*\\)" 1 font-lock-string-face)

     ;; --- Email addresses ---
     '("\\<[a-zA-Z0-9._%+-]+@[a-zA-Z0-9.-]+\\.[a-zA-Z]\\{2,\\}\\>" . font-lock-string-face)

     ;; --- Function definition: name: fn/does/closure ---
     '("\\([a-zA-Z_][a-zA-Z0-9\\-?!.*+<>=\\\\]*\\):\\s-+\\(?:fn[1c\\\\]?\\|does\\|closure\\|pfn\\)\\b"
       1 font-lock-function-name-face)

     ;; --- Set-word: word: (assignment) ---
     '("\\([a-zA-Z_][a-zA-Z0-9\\-?!.*+<>=\\\\]*\\):" 1 font-lock-variable-name-face)

     ;; --- Mod-word: word:: (reassignment) ---
     '("\\([a-zA-Z_][a-zA-Z0-9\\-?!.*+<>=\\\\]*\\)::" 1 font-lock-variable-name-face)

     ;; --- Left mod-word: ::word (injection) ---
     '("::\\([a-zA-Z_][a-zA-Z0-9\\-?!.*+<>=\\\\]*\\)" 1 font-lock-variable-name-face)

     ;; --- Left set-word: :word (injection) ---
     '(":\\([a-zA-Z_][a-zA-Z0-9\\-?!.*+<>=\\\\]*\\)" 1 font-lock-variable-name-face)

     ;; --- Return prefix: ^word ---
     '("\\^\\([a-zA-Z][a-zA-Z0-9\\-?!.*+<>=\\\\]*\\)" 0 font-lock-warning-face)

     ;; --- Get-word: ?word ---
     '("\\?\\([a-zA-Z_][a-zA-Z0-9\\-?!.*+<>=\\\\]*\\)" 0 font-lock-variable-name-face)

     ;; --- Tag-word: 'word ---
     '("'\\([a-zA-Z_][a-zA-Z0-9\\-?!.*+<>=\\\\]*\\)" 0 font-lock-constant-face)

     ;; --- XWord: <word> (type tags) ---
     '("<\\([a-zA-Z][a-zA-Z0-9\\-]*\\)>" 0 font-lock-type-face)

     ;; --- Closing XWord: </word> ---
     '("</\\([a-zA-Z][a-zA-Z0-9\\-]*\\)>" 0 font-lock-type-face)

     ;; --- Op-word: .word (method call) ---
     '("\\.\\([a-zA-Z_][a-zA-Z0-9\\-?!.*+<>=\\\\]*\\)" 0 font-lock-function-name-face)

     ;; --- Pipe-word: |word (pipe call) ---
     '("|\\([a-zA-Z_][a-zA-Z0-9\\-?!.*+<>=\\\\]*\\)" 0 font-lock-function-name-face)

     ;; --- Context path: word/word ---
     '("\\([a-zA-Z_][a-zA-Z0-9\\-?!.*+<>=\\\\]*/[a-zA-Z_][a-zA-Z0-9\\-?!.*+<>=\\\\/]*\\)"
       1 font-lock-type-face)

     ;; --- Constants ---
     (cons constants font-lock-constant-face)

     ;; --- Void/placeholder: _ ---
     '("\\(?:^\\|[[:space:]{}(\\[,]\\)\\(_\\)\\(?:[[:space:]{}()\\],;]\\|$\\)" 1 font-lock-constant-face)

     ;; --- Function definition keywords ---
     (cons kw-fn-def font-lock-keyword-face)

     ;; --- Control flow keywords ---
     (cons kw-control font-lock-keyword-face)

     ;; --- Iteration keywords ---
     (cons kw-iter font-lock-keyword-face)

     ;; --- Error handling keywords ---
     (cons kw-error font-lock-keyword-face)

     ;; --- Context keywords ---
     (cons kw-ctx font-lock-keyword-face)

     ;; --- Combinator keywords ---
     (cons kw-comb font-lock-keyword-face)

     ;; --- Type keywords ---
     (cons kw-types font-lock-type-face)

     ;; --- Other keywords ---
     (cons kw-other font-lock-keyword-face)

     ;; --- Logical operators ---
     (cons logical-ops font-lock-keyword-face)

     ;; --- Builtins ---
     (cons bi-coll font-lock-builtin-face)
     (cons bi-str font-lock-builtin-face)
     (cons bi-print font-lock-builtin-face)
     (cons bi-conv font-lock-builtin-face)
     (cons bi-math font-lock-builtin-face)
     (cons bi-time font-lock-builtin-face)
     (cons bi-io font-lock-builtin-face)
     (cons bi-test font-lock-builtin-face)
     (cons bi-other font-lock-builtin-face)

     ;; --- Numbers ---
     '("\\(?:^\\|[[:space:]{}(\\[,]\\)\\(-?[0-9]+\\.[0-9]+\\)\\b" 1 font-lock-constant-face)
     '("\\(?:^\\|[[:space:]{}(\\[,]\\)\\(-?[0-9]+\\)\\b" 1 font-lock-constant-face)

     ;; --- Operators ---
     '("\\(->\\|<-\\|~>\\|<~\\|=>\\|>=\\|<=\\|>>\\|<<\\|++\\|//\\)" . font-lock-keyword-face)
     '("[[:space:]]\\([+\\-*/%=<>]\\)[[:space:]]" 1 font-lock-keyword-face)

     ;; --- Pipe operator ---
     '("\\(|\\)[[:space:]]" 1 font-lock-keyword-face)

     ;; --- Comma separator ---
     '("," . font-lock-punctuation-face)
     ))
  "Font-lock keywords for `rye-mode'.")

;; ============================================================
;; Indentation
;; ============================================================

(defun rye-indent-line ()
  "Indent the current line as Rye code."
  (interactive)
  (let ((indent (rye--calculate-indent))
        (offset (- (current-column) (current-indentation))))
    (indent-line-to indent)
    (when (> offset 0)
      (forward-char offset))))

(defun rye--calculate-indent ()
  "Calculate the indentation for the current line."
  (save-excursion
    (beginning-of-line)
    (cond
     ;; At the beginning of buffer
     ((bobp) 0)
     ;; Closing bracket on this line — outdent
     ((looking-at "^[[:space:]]*[}\\])]")
      (rye--matching-open-indent))
     ;; Otherwise, look at previous non-empty line
     (t
      (rye--previous-line-indent)))))

(defun rye--matching-open-indent ()
  "Find the indentation of the line with the matching open bracket."
  (save-excursion
    (beginning-of-line)
    (skip-chars-forward " \t")
    (let ((close-char (char-after)))
      (condition-case nil
          (progn
            (forward-char 1)
            (backward-sexp 1)
            (current-indentation))
        (error 0)))))

(defun rye--previous-line-indent ()
  "Calculate indent based on the previous non-empty line."
  (save-excursion
    (forward-line -1)
    (while (and (not (bobp))
                (looking-at "^[[:space:]]*$"))
      (forward-line -1))
    (let ((prev-indent (current-indentation))
          (opens 0))
      ;; Count net opening brackets on previous line
      (beginning-of-line)
      (while (not (eolp))
        (let ((ch (char-after)))
          (cond
           ((memq ch '(?\{ ?\[ ?\()) (setq opens (1+ opens)))
           ((memq ch '(?\} ?\] ?\))) (setq opens (1- opens)))))
        (forward-char 1))
      (max 0 (+ prev-indent (* (max 0 opens) rye-indent-offset))))))

;; ============================================================
;; Keymap
;; ============================================================

(defvar rye-mode-map
  (let ((map (make-sparse-keymap)))
    map)
  "Keymap for `rye-mode'.")

;; ============================================================
;; Mode definition
;; ============================================================

;;;###autoload
(define-derived-mode rye-mode prog-mode "Rye"
  "Major mode for editing Rye programming language files.

\\{rye-mode-map}"
  :syntax-table rye-mode-syntax-table
  :group 'rye

  ;; Comments
  (setq-local comment-start "; ")
  (setq-local comment-end "")
  (setq-local comment-start-skip ";+\\s-*")

  ;; Font lock
  (setq-local font-lock-defaults '(rye-font-lock-keywords nil nil))
  (setq-local font-lock-multiline t)

  ;; Indentation
  (setq-local indent-line-function #'rye-indent-line)
  (setq-local tab-width rye-indent-offset)
  (setq-local indent-tabs-mode nil)

  ;; Electric pairs
  (setq-local electric-pair-pairs
              '((?\{ . ?\})
                (?\[ . ?\])
                (?\( . ?\))
                (?\" . ?\")
                (?\` . ?\`)))

  ;; Paragraph
  (setq-local paragraph-start "\\s-*$")
  (setq-local paragraph-separate "\\s-*$")

  ;; Navigation
  (setq-local beginning-of-defun-function #'rye-beginning-of-defun)
  (setq-local end-of-defun-function #'rye-end-of-defun))

(defun rye-beginning-of-defun (&optional arg)
  "Move backward to the beginning of a Rye function definition.
ARG is the number of definitions to move back."
  (interactive "^p")
  (or arg (setq arg 1))
  (re-search-backward "^[a-zA-Z_][a-zA-Z0-9\\-?!.*+<>=\\\\]*:\\s-" nil 'move arg))

(defun rye-end-of-defun (&optional arg)
  "Move forward to the end of a Rye function definition.
ARG is the number of definitions to move forward."
  (interactive "^p")
  (or arg (setq arg 1))
  (if (re-search-forward "^[a-zA-Z_][a-zA-Z0-9\\-?!.*+<>=\\\\]*:\\s-" nil 'move arg)
      (beginning-of-line)
    (goto-char (point-max))))

;; ============================================================
;; File association
;; ============================================================

;;;###autoload
(add-to-list 'auto-mode-alist '("\\.rye\\'" . rye-mode))

(provide 'rye)

;;; rye.el ends here
