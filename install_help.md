# Offline installation of MySQL and HAProxy

## MySQL RPM installation

1. Download `wget https://dev.mysql.com/get/mysql57-community-release-el7-11.noarch.rpm`
2. Install MySQL source `sudo yum localinstall mysql57-community-release-el7-11.noarch.rpm`
3. Check whether the MySQL source is installed successfully

    ```bash
    [vagrant@bogon ~]$ sudo yum repolist enabled | grep "mysql.*-community.*"
    mysql-connectors-community/x86_64 MySQL Connectors Community 118
    mysql-tools-community/x86_64 MySQL Tools Community 95
    mysql57-community/x86_64 MySQL 5.7 Community Server 364
    ```
4. Download all dependent packages to the local directory (vagrant centos7)
    * On vagrant centos7: install the plugin `sudo yum install yum-plugin-downloadonly`
    * On vagrant centos7: Download dependency packages `sudo yum install -y --downloadonly --downloaddir=/vagrant/mysql57 mysql-community-server`
        > [root@BJCA-device ~]# yum -h
        > -y, --assumeyes answer yes to all questions
        > --downloadonly download only without updating
        > --downloaddir=DLDIR specifies an additional folder to store packages
    * Only keep the rpm packages starting with mysql, delete the rest of the rpm packages

        ```bash
        [vagrant@bogon mysql-community-server-5.7.27-1.el7.x86_64]$ ls -lh *.rpm
        -rw-r--r-- 1 vagrant vagrant  25M July  18 10:59 mysql-community-client-5.7.27-1.el7.x86_64.rpm
        -rw-r--r-- 1 vagrant vagrant 275K July  18 10:59 mysql-community-common-5.7.27-1.el7.x86_64.rpm
        -rw-r--r-- 1 vagrant vagrant 2.2M July  18 11:00 mysql-community-libs-5.7.27-1.el7.x86_64.rpm
        -rw-r--r-- 1 vagrant vagrant 2.1M July  18 11:00 mysql-community-libs-compat-5.7.27-1.el7.x86_64.rpm
        -rw-r--r-- 1 vagrant vagrant 166M July  18 11:00 mysql-community-server-5.7.27-1.el7.x86_64.rpm
        ```
    * Local: upload directory `sshpass -p mima scp -P1122 -o StrictHostKeyChecking=no ./*.rpm root@192.168.1.23:./mysql/`
    * Target machine: execute `sudo rpm -ivh *.rpm --nodeps --force`

        ```bash
        [vagrant@bogon mysql-community-server-5.7.27-1.el7.x86_64]$ sudo rpm -ivh *.rpm --nodeps --force
        Warning: mysql-community-client-5.7.27-1.el7.x86_64.rpm: header V3 DSA/SHA1 Signature, key ID 5072e1f5: NOKEY
        Preparing... ################################# [100%]
        Upgrading/installing...
        1:mysql-community-common-5.7.27-1.e####################################################################################################################################################################################################################################################################################################################################################################################################################################################################################################################################################################### ]
        2:mysql-community-libs-5.7.27-1.el7################################## [ 40 %]
        3:mysql-community-client-5.7.27-1.e################################## [ 60% ]
        4:mysql-community-server-5.7.27-1.e################################## [ 80% ]
        5:mysql-community-libs-compat-5.7.2################################## [100%]
        ````

    * Target machine: start `systemctl enable mysqld`, start the service `systemctl start mysqld`, check the status `systemctl status mysqld`

1. Modify the root local account password
    After the installation is complete, the default password generated is in the /var/log/mysqld.log file. Use the `grep 'temporary password' /var/log/mysqld.log`
    command to find the password in the log.

    ```sql
    ALTER USER 'root'@'localhost' IDENTIFIED BY 'A1765527-61a0';
    ```

    > Note: The password security check plugin (validate_password) is installed by default in mysql 5.7. The default password check policy requires that the password must contain: uppercase and lowercase letters, numbers and special symbols, and the length cannot be less than 8 characters.
    > Otherwise, ERROR 1819 (HY000): Your password does not satisfy the current policy requirements will be prompted.

    ```sql
    use mysql;

    update user set host='localhost' where user='root';
    flush privileges; -- Only allow root to log in locally

    update user set host='%' where user='root';
    flush privileges; -- Allow root remote access
    ```

Remarks:

1. `MYSQL_PWD=\!QAZ2wsx mysql -S /usr/local/mysql/data/mysql.sock -uroot -P13306`, default /tmp/mysql.sock

Thanks,