# symbolassert
A package that can assert if independent defined values are equal.

This package was born to use a test to verify the definition of OS dependent constants and types being equal while not referring to some shared package.

## inspiration
The strconv package has it's own implementation that determines if a charachter is printable and does not import the package that can provide the canonical answer, the unicode package. However there is a test that verifies that the function in strconv agrees with the canonical one in the unicode package. By doing this there is no need to import all the unicode tables if a program wants to use the strconv package.
