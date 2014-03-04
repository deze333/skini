INI parsing in Go
-----------------
Minimalistic `.ini` file parser for golang. Light enough to paste directly into the code.

Understands files like:
```ini
# Web server
server.name = MyServer
server.port = :8080

# DB server
db.server = localhost
db.collection = test

# Some texts
text.hello = Hello, this = sign will not break parsing
```
###Limitations:
* Sections are not supported
* Multi line list values are not supported
