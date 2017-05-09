# Data browser for X3 Albion Prelude and mods #

## Why? ##

The problem with many mods to X3 is that they are often not very well
documented and even if there at some point were various web pages with
maps, lists of ships, weapon stats, etc. they have bit-rotted and are
no longer available or out-of-date. Things get even worse when you
start combining mods, then you can be sure that whatever information
exists is wrong.

This is silly. All the information various sites provide is fully
available on every installation and is perfectly accurate for that
installation. There is no reason to try to figure out which scraps of
information are out there that match your installation because you
have all that information yourself.

## What? ##

This is a simple program that reads your installation of AP including
all mods, extracts all the information it can (minus the stuff that
hasn't been implemented yet, which is most of it) then it then gives
you that information through a webserver that's running on your own
machine which you can access like any other site out there in the
world. The site will never go down unless you want to, it will never
be out-of-date.

## How? ##

I'm a unix person, so this is a command line thing. You start cmd
and type something like this:

      c:\whereever\you\installed\this\x3t.exe "c:\whereever\you\installed\x3"

Then you take your browser (on the same machine) and go to
http://localhost:8080/

That's it.

There is one option. You can add `-listen <host>:<port>` after x3t.exe
to pick which host:port combination the internal web server should
listen to. Specify `localhost:4711` if you want to change the listen
port to 4711. Specify `:8080` if you want it to be accessible to the
world. There are no other options, this program doesn't have any
settings, it doesn't save anything on your computer, it is not
configurable in any way whatsoever. All content that is not extracted
from your x3 installation is built in and static.

## What? (2) ##

In this first proof-of-concept version only a few things are
implemented. A map, basic sector information, a semi-searchable ship
list and some ship details. If there's no interest, I might not bother
implementing more.

## Why is it so ugly? ##

I'm no web designer. Figuring out the data formats and stuff in X3 is
fun, making things pretty isn't. The information is there. Deal with
it or fix it and send pull requests on github.

## How do I build and change this? ##

If you're inclined to work on this, you need Go. I have worked on
things mostly with Go 1.6.3 and 1.8. With the Go 1 compatibility
guarantee there is no reason for things to not keep working with
future versions.  All dependency libraries are vendored, except
one. To build you'll need `https://github.com/jteeuwen/go-bindata`

Install `go-bindata`, then go to our source directory, type `go
generate` and then `go build x3t`.

`go generate` will use `go-bindata` for embedding all the various css,
javascript and some images into the binary so that the distribution of
this program can be one file and not a massive installer that spreads
its tentacles all over your filesystem.

The layout of the source tree:

 * main.go - general setup of everything

 * ships.go, map.go - functionality specific to presenting ships and
   the map. Complete mess at this moment. Things are not where they
   should be and there are too many unnecessary dynamic funcs for
   templates.

 * main_test.go - woefully inadequate test. It's just one benchmark I
   used when trying to figure out how to parse x3_universe.xml faster
   (it's 0.5s right now which is way too slow, but I can't make it
   better without putting an unreasonable amount effort into it).

 * xt/ - package for accessing x3 data. Undocumented and testless. Can
   only read the files.  Write support won't happen until someone
   explains to me what the actual rules are for scrambling the various
   gzip files. Currently I just calculate the xor cookies on the fly.
   Ask me if you want to understand what's going on under the hood and
   how to change stuff. There's a lot of interesting stuff going on
   here and it was quite fun to write.

   * xt/x.go - container to access everything from one x3 installation

   * xt/xfiles.go - decoding of cat/dat files, decoding of pck files and
     rules for which files override what (as I understand them).

   * xt/text.go - access to `t/` text files. I have no idea if I followed
     the official rules, but it seems to work and I'm not getting any
     missing strings.

   * xt/tparse.go - parser primarily for `types/*.txt` should be fixed up
     for other files with a similar format.

   * xt/types.go - structures to contain parsed data from `types/`.

   * xt/universe.go - parser and data structures for x3_universe.xml.

   * xt/extra.go - dumping ground for hardcoded things I couldn't
     figure out how to extract from the game files.

 * icons/ - material.io icons for the map. No longer used (they are
   manually inlined into the big svg), but they are kept around for
   the time when I decide to properly include them instead of just the
   current cut 'n paste method.

 * xtool/ - separate program for acessing x3 data. Has three
   sub-commands - `ls` to list all the files, `cat` to print a file,
   `grep` to grep for a string in all the files. Very crude, but
   useful for debugging.

 * vendor/github.com/lukegb/dds/ - vendored package for reading dds
   files. Except that it's nothing like the actual package because
   I had to rewrite the meat of it to actually work. Also, the dds
   files aren't used anywhere except for a temporary `/foo.png` url
   just to test that it works (as far as I remember it contains
   the image with various map icons).

 * assets/ - all the stuff that gets inlined by `go-bintool`.

  * assets/static - all the files in here are directly accessible from
   the http server under `/static/`. So there's javascript, css and
   some style images.

  * assets/templates - templates for the various pages. They are all
   standard go html/template.

   * all - header and footer for all pages

   * about - dumping ground for licenses and such

   * map - the map. What you get when you point your browser to `/map`

   * map-sector - one sector of the map (the square and all the stuff
     in it).

   * sector - What you get when you point your browser to `/sector/x/y`

   * ship - What you get at `/ship/Name`

   * ships - cat pictures

## template funcs ##

To get things working, there are a bunch of funcs provided for the
templates.

`calc` - since go templates don't do any arithmetic whatsoever this is
an RPN calculator to do simple math. For example, to calculate the
shield strength on a ship in MJ we do:

    {{calc .ShieldType.Strength .MaxShieldCount "*" 1000 "/"}}

`sectName` - gives us the name of a sector. This should have been done
automagically on loading, but I never got around to fixing that. Maybe
this will make sense if we want to change languages (things are
hard-coded to 44 right now).

`raceName` - race number to human-readable string

`lnBreak` - take a string with spaces, break it up into substrings
with certain maximum length. Used for rendering the sector name text
in `map-sector`

`sectorIcons` - take a Sector, generate the icons it should have.  A
mountain if it has more than 600 ore (yes, this was developed on LU).
A "chip" if it has more than 300 wafers. A ship if capitals can dock
somewhere. A sun if there's more than 150% sun in the sector.

`validGate` - take a Gate and see if it leads somewhere.

`asteroidType` - asteroid number to a human-readable string.

`sunPercent` - correctly(?) calculate the percentage of sun in a sector.

`maskToLasers` - take a laser mask and wareclass of a ship, return an
array of lasers that match.

`cockpitPos` - cockpit nubmer to human-redable string.

`shipClassName` - human-readable ship class.

`countGuns` - count the guns on a ship. Because go templates can't
change variables.

`shipClassList` - list of the ship classes we care about.

`isChecked` - helper function to figure out if a checkbox was checked
in a form.

`raceList` - races we care about.


## TODO ##

 - Legend for the map

 - Clean up funcs, mostly things out so that stuff lives in `xt.X` if it
   belongs there. Make some other stuff live in `state`. Make others
   generic functions.

 - Consistent naming of structs. Probably prefix everything from universe
   with U and types with T.

 - Figure out on-the-fly reasonable amounts of ore and wafers for the
   icons. Let's say that top 10% systems should get the icons.

 - More information about ships.

 - More information about sectors

 - Ship configurator. This is very relevant in LU with the sustain
   problems for laser energy. Should be pretty trivial, drop down for
   each laser slot. Calculate energy consumption of the lasers per
   second, subtract laser energy regeneration, divide total laser energy
   by that and we know how long we can sustain fire.

 - Missiles

 - Vendor search, I always forget half the places that sell jumpdrives
   in early AP playthroughs. Yesterday I started a new AP playthrough
   after doing mods for months and I flew straight to Legend's Home to
   get a jumpdrive because I just forgot that Home of Light exists.
   Btw. who sold Drakes? It's just not in my head anymore.

 - Parse IconData.txt and use that to generate icons. Because we can.

 - Draw the gate connections on the map.

 - Remove unnecessary stuff from `state`. In fact, it might not even
   need to exist at all anymore.

 - Figure out how many ships of what type can dock on various ships
   and docks. This was a terrible rabbit hole last time I dug into it,
   but it should be possible to figure out.

 - Render ship models with WebGL? Because why not. Shouldn't be that
   hard. I want to poke at WebGL some time anyway.