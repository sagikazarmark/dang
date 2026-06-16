; Topiary formatting queries for Dang.
;
; The Go formatter in pkg/dang/format.go is still the source of truth for
; semantic rewrites that Topiary cannot express, such as import sorting. These
; queries mirror its whitespace/layout rules for tree-sitter-driven formatting.

; Preserve token interiors that carry user-authored text or lexical structure.
[
  (comment_token)
  (doc_string)
  (string)
  (triple_quote_string)
  (single_template)
  (multi_template)
] @leaf

; Blank lines are stripped unless they appear before comments or declaration
; boundaries where the Go formatter also preserves/introduces separation.
[
  (comment_token)
  (object_decl)
  (interface_decl)
  (union_decl)
  (enum_decl)
  (scalar_decl)
  (directive_decl)
  (new_constructor_decl)
] @allow_blank_line_before

; Comments keep their input-relative position; line comments always terminate
; the line.
(comment_token) @prepend_input_softline @append_hardline

; Top-level forms are line-oriented and files end with a trailing newline.
(dang) @append_hardline

(dang
  [
    (import)
    (decl)
    (reassignment)
    (form)
  ] @append_hardline
)

; Public visibility is implicit for typed declarations/methods. Keep `pub` on
; value-only fields, because deleting it would reparse the line as assignment.
[
  (type_and_args_and_block_field vis: (visibility (pub_token) @delete))
  (type_and_args_field vis: (visibility (pub_token) @delete))
  (type_and_block_field vis: (visibility (pub_token) @delete))
  (type_and_value_field vis: (visibility (pub_token) @delete))
  (type_only_field vis: (visibility (pub_token) @delete))
  (type_only_fun_field vis: (visibility (pub_token) @delete))
]

; Keyword spacing.
[
  (break_token)
  (case_token)
  (catch_token)
  (continue_token)
  (directive_token)
  (else_token)
  (enum_token)
  (if_token)
  (implements_token)
  (import_token)
  (interface_token)
  (let_token)
  (new_token)
  (on_token)
  (pub_token)
  (raise_token)
  (return_token)
  (scalar_token)
  (try_token)
  (type_token)
  (union_token)
] @append_space

[
  (else_token)
  (catch_token)
  (implements_token)
] @prepend_space

; Operators and infix punctuation.
[
  (additive_op)
  (ampersand_token)
  (and_token)
  (arrow_token)
  (double_colon_token)
  (double_interro_token)
  (equality_op)
  (equal_token)
  (multiplicative_op)
  (or_token)
  (plus_equal_token)
  (relational_op)
  "="
  "|"
] @prepend_space @append_space

(colon_token) @prepend_antispace @append_space

[
  (dot_token)
  (immediate_bracket)
  (immediate_paren)
  ")"
  "]"
] @prepend_antispace

[
  (dot_token)
  (immediate_bracket)
  (immediate_paren)
  "("
  "["
] @append_antispace

; A postfix ! must stay attached to the type/expression, but it must not cancel
; spaces required by a following operator such as `=`.
(bang_token) @prepend_antispace

; Commas are followed by a space by default. Multi-line containers/calls add
; hardlines in context-specific rules below, because softlines on comma_token
; would otherwise scope to the immediate sep node rather than the container.
(comma_token) @prepend_antispace @append_space

; Preserve the Go formatter's leading-dot style when a chain was written over
; multiple lines; keep single-line selections tight.
(dot_token) @prepend_input_softline

(select_or_call
  (dot_token) @prepend_indent_start
) @append_indent_end

; Enum members are space-separated in single-line enum declarations.
(enum_decl
  (caps_symbol) @append_space
  .
  (caps_symbol)
)

; Argument lists.
(
  [
    (arg_values . (immediate_paren) @append_begin_scope ")" @prepend_end_scope .)
    (arg_types . (immediate_paren) @append_begin_scope ")" @prepend_end_scope .)
  ]
  (#scope_id! "args")
)

[
  (arg_values . (immediate_paren) @append_empty_softline @append_indent_start)
  (arg_types . (immediate_paren) @append_empty_softline @append_indent_start)
]

[
  (arg_values ")" @prepend_empty_softline @prepend_indent_end .)
  (arg_types ")" @prepend_empty_softline @prepend_indent_end .)
]

; Lists and type application keep single-line input on one line, but indent when
; the user split the construct over lines.
[
  (list . "[" @append_empty_softline @append_indent_start)
  (applied_type . (named_type) (immediate_bracket) @append_empty_softline @append_indent_start)
]

[
  (list "]" @prepend_empty_softline @prepend_indent_end .)
  (applied_type "]" @prepend_empty_softline @prepend_indent_end .)
]

(list
  (_) @append_delimiter
  .
  (sep (comma_token)? @do_nothing)
  (#delimiter! ",")
  (#multi_line_scope_only! "list")
)

(list
  (sep (comma_token) @append_spaced_scoped_softline)
  (#scope_id! "list")
)

(list
  last: (_) @append_delimiter
  (#delimiter! ",")
  (#multi_line_scope_only! "list")
)

(arg_values
  (argument) @append_delimiter
  .
  (sep (comma_token)? @do_nothing)
  (#delimiter! ",")
  (#multi_line_scope_only! "args")
)

(arg_values
  (sep (comma_token) @append_spaced_scoped_softline)
  (#scope_id! "args")
)

(arg_values
  last: (argument) @append_delimiter
  (#delimiter! ",")
  (#multi_line_scope_only! "args")
)

(arg_types
  (arg_type) @append_delimiter
  .
  (sep (comma_token)? @do_nothing)
  (#delimiter! ",")
  (#multi_line_scope_only! "args")
)

(arg_types
  (sep (comma_token) @append_spaced_scoped_softline)
  (#scope_id! "args")
)

(arg_types
  last: (arg_type) @append_delimiter
  (#delimiter! ",")
  (#multi_line_scope_only! "args")
)

(list
  .
  "[" @append_begin_scope
  "]" @prepend_end_scope
  .
  (#scope_id! "list")
)

; Object literals/selections use Dang's double-brace syntax.
[
  (object_literal . "{{" @append_empty_softline @append_indent_start)
  (object_selection . "{{" @append_empty_softline @append_indent_start)
  (object_type . "{{" @append_empty_softline @append_indent_start)
]

[
  (object_literal "}}" @prepend_empty_softline @prepend_indent_end .)
  (object_selection "}}" @prepend_empty_softline @prepend_indent_end .)
  (object_type "}}" @prepend_empty_softline @prepend_indent_end .)
]

; Schema declarations always render their bodies as proper blocks.
[
  (object_decl block: (block . "{" @prepend_space @append_hardline @append_indent_start))
  (interface_decl block: (headers_block . "{" @prepend_space @append_hardline @append_indent_start))
]

[
  (object_decl block: (block "}" @prepend_hardline @prepend_indent_end .))
  (interface_decl block: (headers_block "}" @prepend_hardline @prepend_indent_end .))
]

(headers_block
  .
  "{" @append_hardline @append_indent_start
  (_)
  "}" @prepend_hardline @prepend_indent_end
  .
)

; Block arguments are expressions, so single-line block args stay compact.
[
  (call b: (block_arg . "{" @prepend_space))
  (symbol_block block_arg: (block_arg . "{" @prepend_space))
  (select_or_call name: (field_id) b: (block_arg . "{" @prepend_space))
]

(block_arg
  .
  "{" @append_spaced_softline @append_indent_start
  (_)
  "}" @prepend_spaced_softline @prepend_indent_end
  .
)

(block_params
  (block_param) @append_space
  .
  (block_param)
)

; Enum declarations follow the Go formatter's single-line vs multi-line split.
(enum_decl
  .
  (enum_token)
  (symbol)
  "{" @prepend_space @append_spaced_softline @append_indent_start
  (_)
  "}" @prepend_spaced_softline @prepend_indent_end
  .
)

; Case/try blocks are always line-oriented.
[
  (case "{" @append_hardline @append_indent_start)
  (case "}" @prepend_hardline @prepend_indent_end .)
  (try_catch "{" @append_hardline @append_indent_start)
  (try_catch "}" @prepend_hardline @prepend_indent_end .)
]
