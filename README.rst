======
reddup
======
**Reddup** is a program for cleaning up unused files. The user specifies what
directory they would like to scan for files and how many bytes of files they
would like to clean up. There are commands for listing suggested files to be
cleaned up and for moving those files to another directory.

**Reddup** suggests files to be cleaned up based on their size and when they
were last accessed. Larger files which have not been accessed for a while are
suggested before smaller files which have been accessed recently. The program
also scans for duplicate files. When it finds duplicate files, all copies of
the file except the newest are automatically suggested. Duplicate files don't
count toward the user-defined size limit.

Exclude Patterns
================
**Reddup** supports using shell globbing patterns to exclude files and
directories from being suggested. These patterns have the following format:

* Lines starting with a hash symbol '#' serve as comments.
* An asterisk '*' matches anything, but stops at slashes.
* A double asterisk '**' matches anything, including slashes. Double asterisks
  must be separated from the rest of the pattern by slashes.
* A question mark '?' matches any single character.
* A set of brackets '[]' matches any one of the characters contained within the
  brackets.
* A backslash '\' can be used to escape any of the above meta-characters.
* Patterns starting with a slash match file paths relative to the root of the
  directory that is being searched.
* Patterns not starting with a slash match the ends of file paths anywhere in
  the tree. This is the equivalent of starting the pattern with a double
  asterisk.
