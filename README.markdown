Samay 
=====

Command line Time tracking and reporting
----------------------------------------

Why?
----

* I never find time tracker that I like to use.(classical right?)
* Terminal is always open when my laptop is open.
* I wanted to learn Go(programming language)
* Also wanted to learn protocol buffers.

So, here it is.

Command line based time tracker built with Go(programming language) and protocol buffers. 

Samay?
------

Samay is a hindi word for Time.


Unique features (not so much in reality)
----------------------------------------

* Command line based simple interface
* Uses simple files to store data.
* Store files in Dropbox(detects dropbox folder)
* Reasonably fast.
* Basic reporting (monthly)

Examples
--------

### Start/Stop timer

	$ samay start -p "Project Name"
	$ samay start "Project Name"
	$ samay start MyProject 
	...
	$ samay stop -p "Project Name" -m "worked on Samay Readme"
	$ samay stop MyProject 
		
if you don't specify -m it will open your $EDITOR to enter message 
and close editor to finish.

### Directly log hours

	$ samay entry -p "Project Name" -d 1.5h -m "worked on Samay examples"

A duration string is a possibly signed sequence of decimal numbers, 
each with optional fraction and a unit suffix, such as "300m", "1.5h" 
or "2h45m". Valid time units are "s", "m", "h".

	$ samay entry MyProject -d 30m

Twitter style #hashtags are supported in log message, example message:
"Some time spent in #project #management" that will create two tags
"project" and "management" for that entry
	
This will open editor window to enter log message.


### Reporting
	
	$ samay report 

	Report for April 2013
	------------------------------
	|  Project  |  Hours | Clock |
	------------------------------
	| Samay     |  21:12 |       |
	| SY        |  12:06 |       |
	| Beehive   |   3:16 |       |
	| Chore     |   1:49 |  0:14 |
	------------------------------

	$ samay report -r 3 
	(here 3 is month e.g March)

### Reporting with web interface

	$ samay web

This will open system web browser (on mac os x using `open`) 
and show all projects in tabbed interface.

[![Web interface 1][]][] [![Web interface 2][]][] [![Web interface 3][]][]

  [Web interface 1]: http://farm9.staticflickr.com/8398/8695518062_bf79383b0e_m.jpg
  [![Web interface 1][]]: http://www.flickr.com/photos/niket/8695518062/
    "Web interface 1 by Niket Patel, on Flickr"
  [Web interface 2]: http://farm9.staticflickr.com/8257/8695518054_5b81899b83_m.jpg
  [![Web interface 2][]]: http://www.flickr.com/photos/niket/8695518054/
    "Web interface 2 by Niket Patel, on Flickr"
  [Web interface 3]: http://farm9.staticflickr.com/8537/8695518044_748e7073cd_m.jpg
  [![Web interface 3][]]: http://www.flickr.com/photos/niket/8695518044/
    "Web interface 3 by Niket Patel, on Flickr"


### To add non billable hours

	$ samay stop -p "Project Name" -m "worked on Samay Readme" -bill false

or

	$ samay entry MyProject -d 1.5h -bill false

You can always skip -m option it will open $EDITOR

### Remove project with all of its data

	$ samay remove -p "Project Name"
	$ samay remove MyProject

### Log last 10 entries from Project

	$ samay log -p Samay

or

	$ samay log Samay



Caveats
-------

* If your Dropbox has folder named "Samay" in dropbox root, do not run this utility before renaming that folder to something else.
* This software comes with NO WARRANTIES. 

