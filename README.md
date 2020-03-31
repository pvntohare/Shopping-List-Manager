# Shopping List Application

## Prerequisites
 - install go1.12.5+
 - install redis
 - install mysql@5.7
 - create database ```<dbname>``` in mysql
 - ```mysql -u <username> -p <dbname> < shopping_list_ddl```
 
##Usage
```
go run cmd/webserver/webserver.go [FLAGS] 
FLAGS 
  -db_name shopping_list  specify database name 
  -debug_port 8080        specify port to run debug server on 
  -port 8000              specify port to run this server on
```
## Register user

## User login

## User logout

## Create shopping list

## Share shopping list with other users

## Add items to list

## Mark items from list as Bought/Deleted

## Delete whole list
  
## List registered item categories

