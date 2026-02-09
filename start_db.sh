#!/bin/bash

mkdir -p ./data

docker run -d \
  --name mysql-container \
  -v "$(pwd)/data:/var/lib/mysql" \
  -e MYSQL_ROOT_PASSWORD=secret \
  -p 3306:3306 \
  mysql:latest

echo "MySQL container started with root password 'secret'"
