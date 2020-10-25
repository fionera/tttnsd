# tttnsd
Try taking this nameserver down

# Protocol

## Initial Request
The initial request goes to this address which checks if there is any data.
It returns a String like this:
```
SRV tttnsd.example.com;FEAT FOLDER,HREF,TXT;
```

This instructs the client to use `tttnsd.example.com` as baseurl and which features are supported.

## List
Regex:
```regexp
^(?P<page_number>\d+\.|)(?P<folder_id>|[\.\w+]+|)list.tttnsd.example.com$
```

If there is no page number given, the endpoint returns the amount of pages and items.

```
dig list.tttnsd.example.com
```

```
PAGES 23;ITEMS 2342 
```

---

If there is a page number given, the endpoint returns the given page of the given folder. If the folder id is empty, it returns the root folder. 

Response:
`FD` means that the following is a folder.
`IT` means that the following is an item.

```
FD DIR_NAME|ID;FD DIR_NAME|ID;IT ITEM_NAME|ID;IT ITEM_NAME|ID
```

Example:
```
dig 0.list.tttnsd.example.com
```

```
FD A Folder|EIGH2JOH;FD Another Folder|XAHR9IEM;IT An Item|AEJIE3OP;Another Item|FOE1AYEE
```

## Get Item
Regex: 
```regexp
^(?P<item_id>\w+)((?P<folder_id>|[\.\w+]+)|).tttnsd.example.com$
```

The id enpoint returns the data for the given item_id.

Response:
`00` means that the content is a String
`01` means that the content is a href

```
00 CONTENT
```

Example:
```
item_id = EECHEIS9
folder_ids = AEXAI8AH, EIZA3VAH

dig EECHEIS9.AEXAI8AH.EIZA3VAH.tttnsd.example.com
```

```
00 Hello World!
```