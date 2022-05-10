package mci

import (
	"fmt"
	"os"
	"strings"

	"github.com/bingoohuang/gou/str"
	"github.com/bingoohuang/sqlx"
	"github.com/elliotchance/pie/pie"
	"github.com/sirupsen/logrus"
)

// SlaveStatus contains the slave status information for `show slave status\G`.
type SlaveStatus struct {
	Address              string
	SlaveIOState         string
	MasterHost           string
	SlaveSQLRunningState string
	SlaveIoRunning       string
	SlaveSQLRunning      string
	SecondsBehindMaster  string
	LastIOError          string
}

// CheckMySQLCluster 检查MySQL集群配置.
func (s Settings) CheckMySQLCluster(outputFmt string) {
	if err := s.ValidateAndSetDefault(SetDefault); err != nil {
		logrus.Fatal(err)
	}

	mysqlServerAddrs, err := s.ReadMySQLServersFromHAProxyCfg()
	if err != nil {
		logrus.Fatal(err)
	}

	results := make([]SlaveStatus, 0)

	logrus.Infof("settings:%+v", s)
	logrus.Infof("mysqlServerAddrs:%+v", mysqlServerAddrs)

	// 检查MySQL集群配置时，使用HAProxy配置中的第一条IP
	s.Master1Addr = ""

	pie.Strings(mysqlServerAddrs).Each(func(address string) {
		sepPos := strings.LastIndex(address, ":")
		host, port := address[0:sepPos], address[sepPos+1:]

		if s.Master1Addr == "" {
			s.Master1Addr = host
		}

		s.currentHost = host
		s.Port = str.ParseInt(port)

		db := s.MustOpenGormDB()
		defer db.Close()

		status, err := ShowSlaveStatus(db)
		if err != nil {
			logrus.Fatal(err)
		}

		results = append(results, SlaveStatus{
			Address:              address,
			SlaveIOState:         status.SlaveIOState,
			MasterHost:           status.MasterHost,
			SlaveSQLRunningState: status.SlaveSQLRunningState,
			SlaveIoRunning:       status.SlaveIoRunning,
			SlaveSQLRunning:      status.SlaveSQLRunning,
			SecondsBehindMaster:  status.SecondsBehindMaster,
			LastIOError:          status.LastIOError,
		})
	})

	switch outputFmt {
	case "table":
		TablePrinter{}.Print(results)
	case "json":
		fmt.Println(JSONPretty(results))
	default:
		s.checkMySQLClusterStatus(results)
	}
}

func (s Settings) checkMySQLClusterStatus(results []SlaveStatus) {
	checkResult := ""

	for _, r := range results {
		if r.LastIOError == "" &&
			strings.EqualFold(r.SlaveIoRunning, "Yes") &&
			strings.EqualFold(r.SlaveSQLRunning, "Yes") {
			continue
		}

		checkResult += fmt.Sprintf(
			"Address:%s\nSlaveIoRunning:%s\nSlaveSQLRunning:%s\nLastIOError:%s\n",
			r.Address, r.SlaveIoRunning, r.SlaveSQLRunning, r.LastIOError)
	}

	if checkResult == "" {
		checkResult = "OK"
	}

	fmt.Print(checkResult)
}

// ReadMySQLServersFromHAProxyCfg 检查HAProxy中的MySQL集群配置.
func (s Settings) ReadMySQLServersFromHAProxyCfg() ([]string, error) {
	roConfig, err := SearchFileContent(s.HAProxyCfg, `(?is)mysql-ro(.+)MySQLClusterConfigEnd`)
	if err != nil {
		return nil, fmt.Errorf("searchPatternLinesInFile error %w", err)
	}

	if len(roConfig) == 0 {
		// nolint:goerr113
		return nil, fmt.Errorf("no config found in %s", s.HAProxyCfg)
	}

	lines := str.SplitTrim(roConfig[0], "\n")

	const re = `(?i)^\s*server\s+\S+\s([\w.:]+:\d+)`

	addresses := make([]string, 0)

	for _, line := range lines {
		if strings.HasPrefix(line, "#") {
			continue
		}

		vv, _ := FindRegexGroup1(line, re)
		if len(vv) == 0 {
			continue
		}

		crossIndex := strings.Index(line, "#")
		if crossIndex < 0 {
			addresses = append(addresses, vv[len(vv)-1])
			continue
		}

		commentPart := strings.TrimSpace(line[crossIndex+1:])
		vv, _ = FindRegexGroup1(commentPart, `([\w.:]+:\d+)`)

		if len(vv) >= 1 {
			addresses = append(addresses, vv[0])
		}
	}

	return addresses, nil
}

// CheckMySQL 检查MySQL连接
// refer https://github.com/zhishutech/mysqlha-keepalived-3node/blob/master/keepalived/checkMySQL.py
func (s Settings) CheckMySQL() {
	if s.ValidateAndSetDefault(SetDefault) != nil {
		os.Exit(1)
	}

	psLines, err := Ps([]string{"mysqld"}, []string{"mysqld_safe"}, s.shellTimeoutDuration)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Ps error %v\n", err)
		os.Exit(1)
	}

	if len(psLines) == 0 {
		fmt.Fprintf(os.Stderr, "Ps result is empty\n")
		os.Exit(1)
	}

	fmt.Println(strings.Join(psLines, "\n"))

	pid, cmdName, err := NetstatListen(s.Port)
	if err != nil {
		fmt.Fprintf(os.Stderr, "NetstatListen error %v\n", err)
		os.Exit(1)
	}

	fmt.Printf("netstat found cmd %s with pid %d\n", cmdName, pid)

	if !strings.HasPrefix(cmdName, "mysqld") {
		fmt.Printf("cmd %s is not msyqld\n", cmdName)
		os.Exit(1)
	}

	db := s.MustOpenDB()
	defer db.Close()

	result := sqlx.ExecSQL(db, s.CheckSQL, 100, "NULL")
	if err := PrintSQLResult(os.Stdout, os.Stderr, s.CheckSQL, result); err != nil {
		fmt.Printf("PrintSQLResult error %v\n", err)
		os.Exit(1)
	}
}
