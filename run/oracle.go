package run

import (
	"bufio"
	"database/sql"
	"fmt"
	_ "github.com/sijms/go-ora/v2"
	"github.com/spf13/cobra"
	"golin/global"
	"os"
	"path/filepath"
	"strings"
)

func Oraclestart(cmd *cobra.Command, args []string) {
	//确认结果是否输出
	echotype, err := cmd.Flags().GetBool("echo")
	if err != nil {
		fmt.Println(err)
		return
	}
	echorun = echotype

	//获取分隔符
	spr, err := cmd.Flags().GetString("spript")
	if err != nil {
		fmt.Println(err)
		return
	}
	//如果value值不为空则是运行一次的模式
	value, err := cmd.Flags().GetString("value")
	if err != nil {
		fmt.Println(err)
		return
	}
	if len(value) > 10 {
		Onlyonerun(value, spr, "oracle")
		wg.Wait()
		return
	}
	// 下面开始执行批量的
	ippath, err := cmd.Flags().GetString("ip")
	if err != nil {
		fmt.Println(err)
		return
	}
	//判断pgsql.txt文件是否存在
	Checkfile(ippath, fmt.Sprintf("名称%sip%s用户%s密码%s端口", Split, Split, Split, Split), global.FilePer, ippath)
	// 运行share文件中的函数
	Rangefile(ippath, spr, "oracle")
	wg.Wait()
	//完成前最后写入文件
	Deffile("oracle", count, count-len(errhost), errhost)
}

func OracleRun(name, host, user, passwd, port string) {
	defer wg.Done()
	sid := "orcl"
	oidlist := strings.Split(name, "sid=")
	if len(oidlist) == 2 {
		sid = oidlist[1]
	}
	dsn := fmt.Sprintf("oracle://%s:%s@%s:%s/%s", user, passwd, host, port, sid)
	db, err := sql.Open("oracle", dsn)
	if err != nil {
		errhost = append(errhost, host)
		return
	}

	//确认采集完成目录是否存在
	fullPath := filepath.Join(succpath, "oracle")
	_, err = os.Stat(fullPath)
	if os.IsNotExist(err) {
		os.MkdirAll(fullPath, os.FileMode(global.FilePer))
	}
	firenmame := filepath.Join(fullPath, fmt.Sprintf("%s_%s.log", name, host))
	//先删除之前的同名记录文件
	os.Remove(firenmame)

	//查询版本,如果发生错误肯定输入有问题！
	rows, err := db.Query("select * from v$version")
	if err != nil {
		errhost = append(errhost, host)
		return
	}
	//无问题则创建文件在写入
	file, _ := os.OpenFile(firenmame, os.O_CREATE|os.O_APPEND|os.O_RDWR, os.FileMode(global.FilePer))
	defer file.Close()
	write := bufio.NewWriter(file)
	write.WriteString("--------查询版本\n")
	for rows.Next() {
		var version string
		rows.Scan(&version)
		write.WriteString(version + "\n")
	}

	//退出时写入文件
	write.Flush()

	//是否输出结果
	if echorun {
		readFile, _ := os.ReadFile(firenmame)
		fmt.Println(string(readFile))
	}

}
