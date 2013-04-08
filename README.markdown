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
------------------------------------

* Command line based simple interface
* Uses simple files to store data.
* Store files in Dropbox(detects dropbox folder)
* Reasonably fast.
* Basic reporting (monthly)

Examples
--------

*Start/Stop timer*

	$ samay start -p "Project Name"
	...
	$ samay stop -p "Project Name" -m "worked on Samay Readme"

*Directly log hours*

	$ samay entry -p "Project Name" -d 1.5h -m "worked on Samay examples"

	A duration string is a possibly signed sequence of decimal numbers, 
	each with optional fraction and a unit suffix, such as "300m", "1.5h" 
	or "2h45m". Valid time units are "s", "m", "h".

*Reporting*
	
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

*To add non billable hours*

	$ samay stop -p "Project Name" -m "worked on Samay Readme" -bill false
	- or -
	$ samay entry -p "Project Name" -d 1.5h -m "worked on Samay examples" -bill false

*Remove project with all of its data*

	$ samay remove -p "Project Name"

*Log last 10 entries from Project*

	$ samay log -p Samay

Caveats
-------

* If your Dropbox has folder named "Samay" in dropbox root, do not run this utility before renaming that folder to something else.
* This software comes with NO WARRANTIES. 

