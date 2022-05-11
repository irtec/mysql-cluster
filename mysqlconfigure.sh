#!/usr/bin/env bash

echo "Waiting for mysql to get up"
# Give 30 seconds for master and slave to come up
sleep 15

echo "Create MySQL Servers (master / master repl)"

export MYSQL_PWD=$MYSQL_PWD;

echo "* Create replication user"

mysql -h mysqlmaster1 -uroot -AN -vvv <<InputComesFromHERE
set GLOBAL max_connections=2000;
show variables like "%log_bin%";
CREATE USER '$MYSQL_REPL_USR'@'%';
GRANT REPLICATION SLAVE ON *.* TO '$MYSQL_REPL_USR'@'%' IDENTIFIED BY '$MYSQL_REPL_PWD';
InputComesFromHERE

mysql -h mysqlmaster2 -uroot -AN -vvv <<InputComesFromHERE
set GLOBAL max_connections=2000;
show variables like "%log_bin%";
CREATE USER '$MYSQL_REPL_USR'@'%';
GRANT REPLICATION SLAVE ON *.* TO '$MYSQL_REPL_USR'@'%' IDENTIFIED BY '$MYSQL_REPL_PWD';
InputComesFromHERE

echo "* Set MySQL2 as master on MySQL1"

mysql -h mysqlmaster1 -uroot -AN -vvv <<InputComesFromHERE
STOP SLAVE;
CHANGE MASTER TO master_host='mysqlmaster2', master_port=3306, master_user='$MYSQL_REPL_USR', master_password='$MYSQL_REPL_PWD', MASTER_AUTO_POSITION = 1;
START SLAVE;
InputComesFromHERE

echo "* Set MySQL1 as master on MySQL2"

mysql -h mysqlmaster2 -uroot -AN -vvv  <<InputComesFromHERE
STOP SLAVE;
CHANGE MASTER TO master_host='mysqlmaster1', master_port=3306, master_user='$MYSQL_REPL_USR', master_password='$MYSQL_REPL_PWD', MASTER_AUTO_POSITION = 1;
START SLAVE;
InputComesFromHERE

sleep 3

mysql -h mysqlmaster1 -uroot  -vvv -e "SHOW SLAVE STATUS \G"
mysql -h mysqlmaster2 -uroot  -vvv -e "SHOW SLAVE STATUS \G"

echo "MySQL servers created!"

MASTER1_IP=$(eval "getent hosts mysqlmaster1|awk '{print \$1}'")
MASTER2_IP=$(eval "getent hosts mysqlmaster2|awk '{print \$1}'")

echo $MASTER1_IP       : mysqlmaster1
echo $MASTER2_IP       : mysqlmaster2

mysql -h mysqlmaster1 -uroot -AN -vvv <<InputComesFromHERE
create database bjca;
use bjca;

drop table if exists t1;
CREATE TABLE t1 (
  id int auto_increment,
  a int(11) DEFAULT NULL,
  PRIMARY KEY (id)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;

DROP PROCEDURE IF EXISTS batch_t1;
DELIMITER $
CREATE PROCEDURE batch_t1(
  IN CNT INT
)
BEGIN
  DECLARE i INT DEFAULT 1;
  WHILE i<=CNT DO
    INSERT INTO t1(a) VALUES(i);
    SET i = i+1;
  END WHILE;
END$
DELIMITER ;
InputComesFromHERE
