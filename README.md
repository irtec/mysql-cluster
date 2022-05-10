# docker-compose-mysql-master-master

A docker-compose example for a mysql master master setup

## docker-compose scripts

1. Check the status `docker-compose ps`
2. Forcibly delete the cluster `docker-compose rm -fsv`
3. Start the cluster `docker-compose up`
4. Login No. 1 `docker-compose exec mysqlmaster1 mysql -uroot -proot`
5. Login No. 2 `docker-compose exec mysqlmaster2 mysql -uroot -proot`
6. Stop No. 1 `docker-compose stop mysqlmaster1`
7. Start No. 1 `docker-compose start mysqlmaster1`
8. Check the cluster `./MySQLReplicationCheck.sh`
9. `/bin/bash -x /etc/mysql/conf.d/configure1.sh`
10. `MYSQL_PWD=root mysql -u root`
11. `/tmp/mci -c /tmp/mci.toml`

## testing scenarios

1. Auto-increment field scenario

1. MySQLServer1 inserts 10 pieces of data into t1 table
    1. Check whether the 10 pieces of data in the t1 table in MySQLServer2 are synchronized, and the value of the auto-increment field
    2. MySQLServer2 inserts 10 pieces of data into t1 table
    3. Check whether the 10 pieces of data in the t1 table in MySQLServer1 are synchronized, and the value of the auto-increment field

2. Two-way synchronization scenario
    1. MySQLServer1 inserts 100 pieces of data into t1 table
    2. MySQLServer2 inserts 100 pieces of data into t1 table
    3. Check the synchronization status each, whether it is all 200

3. Node restart scenario
    1. MySQLServer2 stopped
    2. MySQLServer1 inserts 100 pieces of data into t1 table
    3. MySQLServer2 start
    4. View the synchronization status of the t1 table in MySQLServer2

## test SQL scripts

```sql
select name from mysql.proc where name like 't1%';

call bjca.batch_t1(100);
select count(*) from bjca.t1;
select * from bjca.t1;
insert into bjca.t1(a) values(3);
```

## thanks

1. [mysql refman 5.7 Chapter 16 Replication](https://dev.mysql.com/doc/refman/5.7/en/replication.html)
1. [MySQL master slave using docker](https://tarunlalwani.com/post/mysql-master-slave-using-docker/) and its related [github rep](https://github.com/tarunlalwani/docker-compose-mysql-master-slave)
1. [MySQL Master Slave DockerDeployment example](https://chanjarster.github.io/post/mysql-master-slave-docker-example/) and its related [github rep](https://github.com/chanjarster/mysql-master-slave-docker-example)
1. [Play with MySQL dual-master cluster](https://github.com/bingoohuang/blog/issues/118)
1. [High-Availability MySQL cluster with load balancing using HAProxy and Heartbeat.](https://github.com/bingoohuang/docker-compose-mysql-master-master)
1. [Notes on exporting and importing libraries with GTID enabled](https://docs.lvrui.io/2016/10/28/%E5%BC%80%E5%90%AFGTID%E7%9A%84%E6%83%85%E5%86%B5%E4%B8%8B%E5%AF%BC%E5%87%BA%E5%AF%BC%E5%85%A5%E5%BA%93%E7%9A%84%E6%B3%A8%E6%84%8F%E4%BA%8B%E9%A1%B9/)

## tips

### Why do you need to establish a separate docker network

> What was happening was that the default docker network doesn't allow name >> DNS mapping.
> Containers on the default bridge network can only access each other by IP addresses, unless you use the --link option, which is considered legacy. On a user-defined bridge network, containers can resolve each other by name or alias.
>
> --[How to allow docker containers to see each other by their name?](https://serverfault.com/a/913075)

```sql
select * from information_schema.tables where TABLE_SCHEMA not in ('performance_schema', 'information_schema', 'mysql', 'sys') and TABLE_NAME not like '%_mci';
rename table bjca.t2 to bjca.t2_mci;

SET GLOBAL server_id=10002;
STOP SLAVE;
RESET SLAVE ALL;
DROP USER IF EXISTS 'root'@'mysqlmaster1';
CREATE USER 'root'@'mysqlmaster1' IDENTIFIED BY 'root';
GRANT ALL PRIVILEGES ON *.* TO 'root'@'mysqlmaster1' WITH GRANT OPTION;

DROP USER IF EXISTS 'repl'@'%';
CREATE USER 'repl'@'%' IDENTIFIED BY 'repl';
GRANT REPLICATION SLAVE ON *.* TO 'repl'@'%' IDENTIFIED BY 'repl';
CHANGE MASTER TO master_host='mysqlb', master_port=3306, master_user='repl', master_password='repl', master_auto_position = 1;
CHANGE MASTER TO master_host='mysqla', master_port=3306, master_user='repl', master_password='repl', master_auto_position = 1;
START SLAVE;


SET GLOBAL server_id=10001;
STOP SLAVE;
RESET SLAVE ALL;
DROP USER IF EXISTS 'root'@'mysqlmaster1';
CREATE USER 'root'@'mysqlmaster1' IDENTIFIED BY 'root';
GRANT ALL PRIVILEGES ON *.* TO 'root'@'mysqlmaster1' WITH GRANT OPTION;

DROP USER IF EXISTS 'repl'@'%';
CREATE USER 'repl'@'%' IDENTIFIED BY 'repl';
GRANT REPLICATION SLAVE ON *.* TO 'repl'@'%' IDENTIFIED BY 'repl';
CHANGE MASTER TO master_host='mysqlmaster2', master_port=3306, master_user='repl', master_password='repl', master_auto_position = 1;


SHOW MASTER STATUS;
RESET MASTER;
FLUSH TABLES WITH READ LOCK;
SHOW MASTER STATUS;
UNLOCK TABLES;
START SLAVE;


/* main 1 */ create database bjca;
/* main 1 */ create table bjca.t_m1(name varchar(100));
/* main 2 */ insert into bjca.t_m1 values('written from master2');
/* master 1 slave */ select * from bjca.t_m1;

/* main 2 */ create table bjca.t_m2(name varchar(100));
/* main 1 */ insert into bjca.t_m2 values('written from master1');
/* master 2 slave */ select * from bjca.t_m2;

/* all on */ SHOW SLAVE STATUS\G
```

```bash
root@c31810844c58:/# MYSQL_PWD=root mysql -u root
Welcome to the MySQL monitor.  Commands end with ; or \g.
Your MySQL connection id is 3
Server version: 5.7.27-log MySQL Community Server (GPL)

Copyright (c) 2000, 2019, Oracle and/or its affiliates. All rights reserved.

Oracle is a registered trademark of Oracle Corporation and/or its
affiliates. Other names may be trademarks of their respective
owners.

Type 'help;' or '\h' for help. Type '\c' to clear the current input statement.

mysql> SHOW MASTER STATUS;
+------------------+----------+--------------+------------------+------------------------------------------+
| File             | Position | Binlog_Do_DB | Binlog_Ignore_DB | Executed_Gtid_Set                        |
+------------------+----------+--------------+------------------+------------------------------------------+
| mysql-bin.000003 |      836 |              |                  | 7ab075b1-0431-11ea-80f8-0242ac120003:1-8 |
+------------------+----------+--------------+------------------+------------------------------------------+
1 row in set (0.00 sec)

mysql> select TABLE_SCHEMA, TABLE_NAME, from information_schema.tables where TABLE_SCHEMA not in ('performance_schema', 'information_schema', 'mysql', 'sys') and TABLE_NAME not like '%_mci'
    -> ;
+---------------+--------------+------------+------------+--------+---------+------------+------------+----------------+-------------+-----------------+--------------+-----------+----------------+---------------------+---------------------+------------+-----------------+----------+----------------+---------------+
| TABLE_CATALOG | TABLE_SCHEMA | TABLE_NAME | TABLE_TYPE | ENGINE | VERSION | ROW_FORMAT | TABLE_ROWS | AVG_ROW_LENGTH | DATA_LENGTH | MAX_DATA_LENGTH | INDEX_LENGTH | DATA_FREE | AUTO_INCREMENT | CREATE_TIME         | UPDATE_TIME         | CHECK_TIME | TABLE_COLLATION | CHECKSUM | CREATE_OPTIONS | TABLE_COMMENT |
+---------------+--------------+------------+------------+--------+---------+------------+------------+----------------+-------------+-----------------+--------------+-----------+----------------+---------------------+---------------------+------------+-----------------+----------+----------------+---------------+
| def           | bjca         | t2         | BASE TABLE | InnoDB |      10 | Dynamic    |          1 |          16384 |       16384 |               0 |            0 |         0 |           NULL | 2019-11-11 03:17:33 | 2019-11-11 03:17:33 | NULL       | utf8_general_ci |     NULL |                |               |
+---------------+--------------+------------+------------+--------+---------+------------+------------+----------------+-------------+-----------------+--------------+-----------+----------------+---------------------+---------------------+------------+-----------------+----------+----------------+---------------+
1 row in set (0.00 sec)

mysql> rename table bjca.t2 to bjca.t2_mci;
Query OK, 0 rows affected (0.00 sec)

mysql> RESET MASTER;
Query OK, 0 rows affected (0.01 sec)

mysql> FLUSH TABLES WITH READ LOCK;
Query OK, 0 rows affected (0.00 sec)

mysql> SHOW MASTER STATUS;
+------------------+----------+--------------+------------------+-------------------+
| File             | Position | Binlog_Do_DB | Binlog_Ignore_DB | Executed_Gtid_Set |
+------------------+----------+--------------+------------------+-------------------+
| mysql-bin.000001 |      154 |              |                  |                   |
+------------------+----------+--------------+------------------+-------------------+
1 row in set (0.00 sec)
```

How to re-sync the Mysql DB if Master and slave have different database incase of Mysql replication?
https://stackoverflow.com/questions/2366018/how-to-re-sync-the-mysql-db-if-master-and-slave-have-different-database-incase-o

This is the full step-by-step procedure to resync a master-slave replication from scratch:

At the master:

RESET MASTER;
FLUSH TABLES WITH READ LOCK;
SHOW MASTER STATUS;
And copy the values of the result of the last command somewhere.

Without closing the connection to the client (because it would release the read lock) issue the command to get a dump of the master:

mysqldump -u root -p --all-databases > /a/path/mysqldump.sql
Now you can release the lock, even if the dump hasn't ended yet. To do it, perform the following command in the MySQL client:

UNLOCK TABLES;
Now copy the dump file to the slave using scp or your preferred tool.

At the slave:

Open a connection to mysql and type:

STOP SLAVE;
Load master's data dump with this console command:

mysql -uroot -p < mysqldump.sql
Sync slave and master logs:

RESET SLAVE;
CHANGE MASTER TO MASTER_LOG_FILE='mysql-bin.000001', MASTER_LOG_POS=98;
Where the values of the above fields are the ones you copied before.

Finally, type:

START SLAVE;
To check that everything is working again, after typing:

SHOW SLAVE STATUS;
you should see:

Slave_IO_Running: Yes
Slave_SQL_Running: Yes
That's it!

[Daily MySQL 024：FLUSH TABLES](https://blog.csdn.net/qingsong3333/article/details/77170864)

FLUSH TABLES role is flush table, and add the corresponding lock according to the parameters. The default is to write the log. If you do not want to write the log, you can set and add parameters. NO_WRITE_TO_BINLOG。in addition， FLUSH TABLES The commit command is implicitly issued before the command is executed. The common syntax is as follows:

• FLUSH TABLES
Close all tables, including those in use, and flush query cache. If there is an active LOCK TABLES ... READ command is not allowed FLUSH TABLES Order

• FLUSH TABLES tbl_name [, tbl_name] ...
Only FLUSH specified table

• FLUSH TABLES WITH READ LOCK
Close all tables and add a global to all tables in all databases read lock. This is useful for backup operations, after locking, to prevent applications from modifying the database. This is a global level lock, not a table lock.

• FLUSH TABLES tbl_name [, tbl_name] ... WITH READ LOCK
Same as above, but only for some tables

• FLUSH TABLES tbl_name [, tbl_name] ... FOR EXPORT
only for InnoDB Table, can ensure that changes to the table are flushed to disk, MySQL can copy the table by directly copying the underlying file, refer to the link.

[Should a MySQL replication slave be set to read only?](https://dba.stackexchange.com/a/30129)

When a Slave is read-only, it is not 100% shielded from the world.

According to MySQL Documentation on read-only

This variable is off by default. When it is enabled, the server permits no updates except from users that have the SUPER privilege or (on a slave server) from updates performed by slave threads. In replication setups, it can be useful to enable read_only on slave servers to ensure that slaves accept updates only from the master server and not from clients.

Thus, anyone with SUPER privilege can read and write at will to such a Slave...

Make sure all non-privileged users do not have the SUPER Privilege.

If you want to revoke all SUPER privileges in one shot, please run this on Master and Slave:

UPDATE mysql.user SET super_priv='N' WHERE user<>'root';
FLUSH PRIVILEGES;
With reference to the Slave, this will reserve SUPER privilege to just root and prevent non-privileged from doing writes they would otherwise be restricted from.

UPDATE 2015-08-28 17:39 EDT
I just learned recently that MySQL 5.7 will introduce super_read_only.

This will stop SUPER users in their tracks because the 5.7 Docs say

If the read_only system variable is enabled, the server permits client updates only from users who have the SUPER privilege. If the super_read_only system variable is also enabled, the server prohibits client updates even from users who have SUPER. See the description of the read_only system variable for a description of read-only mode and information about how read_only and super_read_only interact.

Changes to super_read_only on a master server are not replicated to slave servers. The value can be set on a slave server independent of the setting on the master.

super_read_only was added in MySQL 5.7.8.

`MYSQL_PWD=root mysqldump -u root -p --all-databases > /a/path/mysqldump.sql`
