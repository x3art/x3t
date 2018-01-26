# xscriptcat

The reason why this tool exists is because I want to be able to get somewhat reasonable diffs in git for x3 MSCI scripts.

The way to set it up is:

 - Have Go installed.

 - `go get -u github.com/x3art/x3t/xscriptcat`, this will install xscriptcat into `$GOPATH/bin`.

 - Create git repository with your scripts.
 
 - In the root of the repository add (or edit) a `.gitattributes` file with the following line:

    addon/scripts/*.xml	diff=xscript

 - In every cloned copy of the repository (this would be a terrible security risk if it happened automatically) edit .git/config and add:

    [diff "xscript"]
        textconv = $HOME/go/bin/xscriptcat.exe

   (Edit the PATH to where the binary was installed)

 - Now every diff of MSCI scripts will show what source code changed instead of the xml mess it is hidden behind.

Indentation is stripped because X-studio that I'm using to edit my scripts seems to pretty much randomize indentation on every edit.