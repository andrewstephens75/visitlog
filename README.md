# VisitLog

A very simple analytics service with no dependencies, written in golang. This is what I run on my [blog/homepage](https://sheep.horse) to keep track of visits. I could use a third-party analytics service but I abhore sending data about my users to random companies.

You can see the data that Visitlog collects: [sheep.horse Visitor Statistics](https://sheep.horse/visitor_statistics.html)

## Goals

* Completely stand alone service - a single executable with no compile time or run time dependencies. Requires no database or thirdparty libraries.
* Very light data gathering. Does not collect information about the user apart from the fact that they visited the page. Does not collect IPs or browser data. No attempt is made to track recurring visits.
* Self-hosted 
* Very simple data structures
* Easily modifiable  
* Safe and secure

## Installation

Compile the project. The `linuxbuild.sh` file provides a simple way to compile using a docker container but compiling natively should be as simple as `go build`.

Copy the executable to wherever you want to run it from. 

Create a file named `visitlogdb` in the same directory. Edit it so that it contains a simple empty json object. 

```
{}
```

This is somewhat of a misfeature. Visitlog should create this file itself on first startup.