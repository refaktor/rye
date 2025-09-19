" Vim syntax file
" Language:	Rye
" Maintainer:	Based on rebol.vim by Gregory Higley <code@revolucent.net>
" Filenames:	*.rye
" Last Change:	2025-07-07
"

" For version 5.x: Clear all syntax items
" For version 6.x: Quit when a syntax file was already loaded
if version < 600
  syntax clear
elseif exists("b:current_syntax")
  finish
endif

" Rye is case sensitive (unlike Rebol)
syn case match

" Define Rye-specific syntax elements
syn match ryeDotMethod "\.\K\k*"
syn match ryeAssignment ":\w\+"
syn match ryeVariableAssign "\w\+::"
syn match ryeVariableRef ":\w\+"
syn match ryeSectionTitle "^section\s\+\".*\"$"
syn match ryeGroupTitle "^group\s\+\".*\"$"

" As per current users documentation
if version < 600
  set isk=@,48-57,?,!,.,',+,-,*,&,\|,=,_,~,^
else
  setlocal isk=@,48-57,?,!,.,',+,-,*,&,\|,=,_,~,^
endif

syn match ryeSheBang "\%^#!.*" display

" Numbers
syn match ryeInteger "\<[+-]\=\d\+\('\d*\)*\>"
syn match ryeDecimal "[+-]\=\(\d\+\('\d*\)*\)\=[,.]\d\=\(e[+-]\=\d\+\)\="
syn match ryeDecimal "[+-]\=\d\+\('\d*\)*\(e[+-]\=\d\+\)\="

" Tuples
syn match	ryeTuple	"\(\d\+\.\)\{2,}"

" Words predefined by Rye at startup.
" This is a partial list based on the examples seen
syn keyword ryePredefined about abs absolute action action? add ajoint all also alter and any any-block? any-function? any-object? any-path? any-string? any-word? append apply aqua arccosine arcsine arctangent array as-pair ascii? ask assert at attempt
syn keyword ryePredefined back backslash backspace base-color beige binary? bind bind? bitset? black block? blue body-of boot-print bound? break brick brown browse bs bugs
syn keyword ryePredefined call case catch cause-error cd change change-dir changes char? charset chat check checksum clean-path clear clos close closure closure? coal coffee collect collect-words command? comment complement compose compress confirm construct context continue copy cosine cr create crimson crlf cyan
syn keyword ryePredefined datatype? datatypes date? debase decimal? decloak decode decode-url decompress default dehex delect delete delete-dir deline delta-profile delta-time demo detab difference dir? dirize divide do do-callback do-codec do-commands docs does dp ds dt dump dump-obj
syn keyword ryePredefined echo eighth either email? empty? enbase encloak encode encoding? enline entab equal? equiv? error? escape even? event? evoke exclude exists? exit exp extend extract
syn keyword ryePredefined fifth file-type? file? find find-all find-script first first+ for forall foreach forest forever form format forskip found? fourth frame? func funco funct function function?
syn keyword ryePredefined get get-env get-path? get-word? gob? gold gray greater-or-equal? greater? green
syn keyword ryePredefined halt handle? has head head? help
syn keyword ryePredefined if image? import in in-dir index? info? input insert integer? intern intersect invalid-utf? issue? ivory
syn keyword ryePredefined join
syn keyword ryePredefined khaki
syn keyword ryePredefined last last? latin1? launch leaf length? lesser-or-equal? lesser? lf lib library? license limit-usage linen list list-dir list-env lit-path? lit-word? load load-extension load-gui log-10 log-2 log-e logic? loop loud-print lowercase ls
syn keyword ryePredefined magenta make make-banner make-dir map map-each map-event map-gob-offset map? maroon max maximum maximum-of min minimum minimum-of mint mkdir mod modified? modify module module? modulo mold mold64 money? more move multiply
syn keyword ryePredefined native native? navy negate negative? new-line new-line? newline newpage next ninth none none? not not-equal? not-equiv? now null number?
syn keyword ryePredefined object object? odd? offset? oldrab olive op? open open? orange or
syn keyword ryePredefined pair? papaya paren? parse past? path? pending percent? pewter pi pick pink poke port? positive? power prin print printf probe protect purple pwd
syn keyword ryePredefined q query quit quote
syn keyword ryePredefined random read rebcode? reblue rebolor recycle red reduce refinement? reflect reform rejoin remainder remold remove remove-each rename repeat repend replace request-file resolve return reverse reword rm round
syn keyword ryePredefined same? save say-browser scalar? script? second secure select selfless? series? set set-env set-path? set-scheme set-word? seventh shift sienna sign? silver sine single? sixth size? skip sky slash snow sort source sp space spec-of speed? split split-path square-root stack stats strict-equal? strict-not-equal? string? struct? subtract suffix? swap switch sys system
syn keyword ryePredefined t tab tag? tail tail? take tan tangent task task? teal tenth third throw time? title-of to to-binary to-bitset to-block to-char to-closure to-command to-datatype to-date to-decimal to-email to-error to-event to-file to-function to-get-path to-get-word to-gob to-hex to-image to-integer to-issue to-lit-path to-lit-word to-local-file to-logic to-map to-module to-money to-object to-pair to-paren to-path to-percent to-port to-rebol-file to-refinement to-relative-file to-set-path to-set-word to-string to-tag to-time to-tuple to-typeset to-url to-vector to-word trace transcode trim true? try tuple? type? types-of typeset?
syn keyword ryePredefined unbind undirize union unique unless unprotect unset unset? until update upgrade uppercase url? usage use utf? utype?
syn keyword ryePredefined value? values-of vector? violet
syn keyword ryePredefined wait wake-up water what what-dir wheat while white why? word? words-of write
syn keyword ryePredefined xor
syn keyword ryePredefined yello yellow
syn keyword ryePredefined zero zero?

" Rye specific keywords
syn keyword ryePredefined fn does section group equal equal\todo var cmd fix fix\either current-ctx parent isolate raw-context private extends bind! unbind! capture-stdout dump doc\of? autotype to-table add-column add-indexes! left-join inner-join group-by
syn keyword ryePredefined map\pos filter seek purge reduce fold partition group produce sum-up rest\from mold\nowrap mold\unwrap doc\of? load\csv table\columns add-col! add-indexes! left-join inner-join group-by order-by! where-equal where-contains where-match where-greater where-lesser where-between
syn keyword ryePredefined vals dict list table type? length? is-integer is-string multiple-of even odd join sort! indexes? header? concat* to-integer to-string
syn keyword ryePredefined args\raw read-all join\with switch either fix\either end newline print-header print-help build-ryel build-fyne install-ryel

" Rye operators
syn keyword ryeOperator ! != !== & * ** + ++ - -- --- / // < <= <> = == =? > >= ? ?? and or xor
syn match ryeOperator "->"
syn match ryeOperator "<-"
syn match ryeOperator ","
syn match ryeOperator "++"

" Rye special refinements
syn match ryeRefinement "\\\K\k*" 

" Rye pipe operator and method chains
syn match ryePipe "|"
syn match ryeMethodChain "\.\K\k*"
syn match ryePipeFunction "|\K\k*"
syn match ryeQuestionMark "?\K\k*"

" Rye special
syn keyword ryeSpecial false off on no none self true yes nil _

" Basics
syn match ryeComment ";.*$"
syn match ryeType "\K\k*!"
syn match ryeRefinementWord "\K\k*" contained
syn match ryeRefinement "/" nextgroup=ryeRefinementWord
syn match ryeGetWord "\K\k*" contained
syn match ryeGet ":" nextgroup=ryeGetWord
syn match ryeLitWord "\K\k*" contained
syn match ryeLit "'" nextgroup=ryeType,ryeLitWord
syn match ryeLocal "/local\>"
syn match ryeSet "\K\k*:"

" Strings
syn match ryeString "\a\+:\/\/[^[:space:]]*" 
syn match ryeString "%[^[:space:]]*"
syn region ryeString oneline start=+%\="+ skip=+^"+ end=+"+ contains=ryeSpecialCharacter
syn region ryeString start=+`+ end=+`+ contains=ryeSpecialCharacter
syn match ryeSpecialCharacter contained "\^[^[:space:][]"
syn match ryeSpecialCharacter contained "%\d\+"

" Blocks
syn region ryeCurlyBlock start="{" end="}" contains=ALL fold
syn region ryeSquareBlock start="\[" end="\]" contains=ALL fold
syn region ryeParenBlock start="(" end=")" contains=ALL fold

com! -nargs=+ RyeHi hi def link <args>

RyeHi ryeComment Comment
RyeHi ryeSheBang Comment
RyeHi ryeOperator Operator
RyeHi ryePipe Operator
RyeHi ryeMethodChain Function
RyeHi ryeDotMethod Function
RyeHi ryeAssignment Identifier
RyeHi ryeVariableAssign Identifier
RyeHi ryeVariableRef Identifier
RyeHi ryePipeFunction Function
RyeHi ryeQuestionMark Function
RyeHi ryeSectionTitle Title
RyeHi ryeGroupTitle Title
RyeHi ryeLocal Special
RyeHi ryeRefinementWord Constant
RyeHi ryeRefinement Constant
RyeHi ryeSpecial Special
RyeHi ryeLitWord Constant
RyeHi ryeLit Constant
RyeHi ryePredefined Keyword
RyeHi ryeInteger Number
RyeHi ryeDecimal Number
RyeHi ryeTuple Number
RyeHi ryeSpecialCharacter Special
RyeHi ryeString String
RyeHi ryeType Type
RyeHi ryeGet Identifier
RyeHi ryeGetWord Identifier
RyeHi ryeSet Identifier
RyeHi ryeCurlyBlock Normal
RyeHi ryeSquareBlock Normal
RyeHi ryeParenBlock Normal

delc RyeHi

syn sync fromstart
let b:current_syntax = "rye"
