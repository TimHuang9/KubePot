docker run -d \
  -p 3306:3306 \
  -e MYSQL_ROOT_PASSWORD=test \
  -v /Users/admin/Documents/code/mysql/logs:/tmp/ \
  -v /Users/admin/Documents/code/mysql/data:/var/lib/mysql \
  -v /etc/localtime:/etc/localtime \
  mysql:8.3.0
