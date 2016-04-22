## Examples

```bash
$ MASSQUERY_CNN="login:password@tcp(books.example.com:3306)/books?allowAllFiles=true" bin/massquery --query "SELECT id, name FROM Books" --format "{status}\t{res}"
```

```bash
$ echo -e "Edgar\tPoe\nHoward\tLovecraft" \ 
    | bin/massquery \
        --cnn "login:password@tcp(books.example.com:3306)/books?allowAllFiles=true" \
        --query "SELECT id, name FROM Books WHERE Author='{1}'" \
        --format "{status}\t{res}"
```

```bash
$ echo -e "\tlogin:password@tcp(books1.example.com:3306)/books?allowAllFiles=true\tEdgar\tPoe\n\tlogin:password@tcp(books2.example.com:3306)/books?allowAllFiles=true\tHoward\tLovecraft" \ 
    | bin/massquery \ 
        --query "SELECT id, name FROM Books WHERE Author='{3}'" 
        --format "{cnn}\t{status}\t{res}"
```

