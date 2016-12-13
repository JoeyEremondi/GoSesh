# GoSesh

The aim of this project is to help increase knowledge about session types from the *Multiparty asynchronous session types* paper at POPL 2008 (Honda et al. 2008). GoSesh provides an accessible way for programmers to increase safety in their distributed systems code using session types and dynamic type checking. Our implementation provides a mockup DSL to model simple session types. While the model is still fairly limited, it provides a solid foundation for more complex session types to be implemented. 

## dynamic

This package contains the dynamic checker that ensures that the current session type matches the specification.

## example

This package contains examples of method stubs and the mockup files from which they are generated.

## godoc

This is the auto-generated html documentation for the library.

## mockup

The DSL for programmers to use when generating mockup files. This is referenced in the *example* folder.

## multiparty

Defines global and local session types used throughout the library.

## test

Internal tests for the library.

# Contributing

1. Fork it!
2. Create a feature branch: `git checkout -b my-new-feature`
3. Commit your changes: `git commit -am 'Add some feature'`
4. Push to the branch: `git push origin my-new-feature`
5. Submit a pull request

# References

Honda, Kohei, Nobuko Yoshida, and Marco Carbone. "Multiparty asynchronous session types." ACM SIGPLAN Notices 43.1 (2008): 273-284.

Neubauer, Matthias, and Peter Thiemann. "Session types for asynchronous communication." Universität Freiburg (2004).
