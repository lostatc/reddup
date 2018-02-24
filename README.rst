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
