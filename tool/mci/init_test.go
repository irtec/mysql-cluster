package mci_test

import (
	"testing"

	"github.com/bingoohuang/tool/mci"
	"github.com/stretchr/testify/assert"
)

const ha = `
listen mysql-rw
  bind 127.0.0.1:13306
  mode tcp
  option tcpka
  server mysql-1 10.0.0.1:3306 check inter 1s
  server mysql-2 10.0.0.2:3306 check inter 1s backup

listen mysql-ro
  bind 127.0.0.1:23306
  mode tcp
  option tcpka
  server mysql-1 10.0.0.1:3306 check inter 1s # 10.0.0.1:3306
  server mysql-2 10.0.0.2:3306 check inter 1s # 10.0.0.2:3306
  server mysql-3 10.0.0.3:3306 check inter 1s # 10.0.0.3:3306
`

func TestMaster1(t *testing.T) {
	settings := &mci.Settings{
		Master1Addr:  "10.0.0.1",
		Master2Addr:  "10.0.0.2",
		Password:     "123456",
		ReplUsr:      "repl",
		ReplPassword: "repl_pwd",
		SlaveAddrs:   []string{"10.0.0.3"},
		Debug:        true,
	}

	result, err := settings.CreateMySQLCluster()
	assert.Nil(t, err)
	assert.Equal(t, []string{
		"DELETE FROM mysql.user WHERE user='repl'",
		"FLUSH PRIVILEGES",
		"CREATE USER 'repl'@'%' IDENTIFIED BY 'repl_pwd'",
		"GRANT REPLICATION SLAVE ON *.* TO 'repl'@'%' IDENTIFIED BY 'repl_pwd'",
		// nolint:lll
		"CHANGE MASTER TO master_host='10.0.0.2',master_port=3306,master_user='repl',master_password='repl_pwd',master_auto_position=1",
	},
		result.Nodes[0].Sqls)
	assert.Equal(t, ha, result.HAProxy)
}

func TestMaster2(t *testing.T) {
	settings := &mci.Settings{
		Master1Addr:  "10.0.0.1",
		Master2Addr:  "10.0.0.2",
		Password:     "123456",
		ReplUsr:      "repl",
		ReplPassword: "repl_pwd",
		SlaveAddrs:   []string{"10.0.0.3"},
		Debug:        true,
	}

	result, err := settings.CreateMySQLCluster()
	assert.Nil(t, err)
	assert.Equal(t,
		[]string{
			"DELETE FROM mysql.user WHERE user='repl'",
			"FLUSH PRIVILEGES",
			"CREATE USER 'repl'@'%' IDENTIFIED BY 'repl_pwd'",
			"GRANT REPLICATION SLAVE ON *.* TO 'repl'@'%' IDENTIFIED BY 'repl_pwd'",
			// nolint:lll
			"CHANGE MASTER TO master_host='10.0.0.1',master_port=3306,master_user='repl',master_password='repl_pwd',master_auto_position=1",
		}, result.Nodes[1].Sqls)
	assert.Equal(t, ha, result.HAProxy)
}

func TestSlave(t *testing.T) {
	settings := &mci.Settings{
		Master1Addr:  "10.0.0.1",
		Master2Addr:  "10.0.0.2",
		Password:     "123456",
		ReplUsr:      "repl",
		ReplPassword: "repl_pwd",
		SlaveAddrs:   []string{"10.0.0.3"},
		Debug:        true,
	}

	result, err := settings.CreateMySQLCluster()
	assert.Nil(t, err)
	assert.Equal(t,
		// nolint:lll
		[]string{"DELETE FROM mysql.user WHERE user='repl'", "FLUSH PRIVILEGES", "CREATE USER 'repl'@'%' IDENTIFIED BY 'repl_pwd'", "GRANT REPLICATION SLAVE ON *.* TO 'repl'@'%' IDENTIFIED BY 'repl_pwd'", "CHANGE MASTER TO master_host='10.0.0.2',master_port=3306,master_user='repl',master_password='repl_pwd',master_auto_position=1"}, result.Nodes[2].Sqls)
}

func TestNone(t *testing.T) {
	settings := &mci.Settings{
		Master1Addr:  "10.0.0.1",
		Master2Addr:  "10.0.0.2",
		Password:     "123456",
		ReplUsr:      "repl",
		ReplPassword: "repl_pwd",
		SlaveAddrs:   []string{"10.0.0.3"},
		Debug:        true,
	}

	result, err := settings.CreateMySQLCluster()
	assert.Nil(t, err)
	assert.Equal(t, 3, len(result.Nodes))
	assert.Equal(t, ha, result.HAProxy)
}
