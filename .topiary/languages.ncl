{
  languages.dang = {
    extensions = ["dang"],
    indent | force = "  ",
    # Build this with: hack/build-topiary-grammar.sh
    grammar.source.path = ".topiary/tree-sitter-dang.so",
  },
}
