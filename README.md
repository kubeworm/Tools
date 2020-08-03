# Tools
Misc Tools

# hidrac.go
As detailed in the comments, this is an admin session tailgating exploit chained to a remote command execution exploit for iDRAC6 devices. also first public CVE-2018-1212 POC?

# CVE-2019-9053-2.py
A proof of concept that the "patch" for CVE-2019-9053 was incomplete.

# CVE-2015-6854.go
A remote memory disclosure exploit in CA SingleSignOn / SiteMinder Agent.

# spraynpray.go
A tool for spraying the php tmp directory with files containing user content, intended to be used with an already known file inclusion vulnerability to obtain arbitrary code execution. Current optimized for musl libc php (I.E. docker alpine).
