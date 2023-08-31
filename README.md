# gopbin
go p(aste)bin

data persistently stored to sqlite

## usage

```http
POST /

content of the paste...

---
HTTP 201

<id of created paste>
```

```http
GET /<id>

content of the paste...
```