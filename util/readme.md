# Utility Functions

Dependency cycles are not allowed in go. Therefore, sometimes as individual packages become very large we arrive at a cycle of dependencies.


Any code that doesn't import anything in packages can go into into this util package,
which should not import any other packages.