package main

import (
	"bufio"
	"flag"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"regexp"
	"strings"
)

/**
%hostname% - where you want to replace to host
%root_dir% - where you want to replace to project root directory
*/
var nginxTemplate = `
server {
	listen 80;
	listen [::]:80;

	root %root_dir%;

	index index.php index.html index.htm index.nginx-debian.html;

	charset utf-8;

	server_name %hostname%;
	error_page 404 /index.php;

	location / {
		try_files $uri $uri/ /index.php?$query_string;
	}

	location ~ \.php$ {
		include snippets/fastcgi-php.conf;
		fastcgi_param SCRIPT_FILENAME $realpath_root$fastcgi_script_name;
		fastcgi_pass unix:/var/run/php/php%php_version%-fpm.sock;
	}
}
`

func main() {
	flag.Parse()

	if *flag.Bool("h", false, "Help") {
		flag.PrintDefaults()
		os.Exit(0)
	}

	args := flag.Args()
	if len(args) == 0 {
		fmt.Println("Please specify command, hint: nxutil (create|enable|disable|remove) [HOSTNAME]")
		os.Exit(1)
	}

	if len(args) < 2 {
		fmt.Println("Please specify host, hint: nxutil (create|enable|disable|remove) [HOSTNAME]")
		os.Exit(1)
	}

	host := args[1]

	switch args[0] {
	case "create":
		if _, err := os.Stat("/etc/nginx/sites-available/" + host); err == nil {
			fmt.Println("Host already exists")
			os.Exit(1)
		}

		var rootDir string

		fmt.Println("Project root directory:")
		for {
			fmt.Print("-> ")
			rd, _ := bufio.NewReader(os.Stdin).ReadString('\n')
			// convert CRLF to LF
			rootDir = strings.Replace(rd, "\n", "", -1)

			if rootDir != "" {
				break
			}
		}

		var phpVersion string

		fmt.Println("Php version, eg \"7.4\" (without quotes):")
		for {
			fmt.Print("-> ")
			rd, _ := bufio.NewReader(os.Stdin).ReadString('\n')
			// convert CRLF to LF
			phpVersion = strings.Replace(rd, "\n", "", -1)

			if phpVersion != "" {
				break
			}
		}

		handledTemplate := strings.Replace(nginxTemplate, "%root_dir%", rootDir, 1)
		handledTemplate = strings.Replace(handledTemplate, "%hostname%", host, 1)
		handledTemplate = strings.Replace(handledTemplate, "%php_version%", phpVersion, 1)

		err := os.WriteFile("/etc/nginx/sites-available/"+host, []byte(handledTemplate), 0666)
		if err != nil {
			log.Fatal(err)
		}

		f, err := os.OpenFile("/etc/hosts",
			os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
		if err != nil {
			log.Println(err)
		}
		defer f.Close()
		if _, err := f.WriteString("\n#host-" + host + "-begin\n127.0.0.1    " + host + "\n#host-" + host + "-end"); err != nil {
			log.Println(err)
		}
	case "enable":
		err := os.Symlink("/etc/nginx/sites-available/"+host, "/etc/nginx/sites-enabled/"+host)
		if err != nil {
			log.Fatal(err)
		}
	case "disable":
		err := os.Remove("/etc/nginx/sites-enabled/" + host)
		if err != nil {
			log.Fatal(err)
		}
	case "remove":
		if _, err := os.Stat("/etc/nginx/sites-enabled/" + host); err == nil {
			err := os.Remove("/etc/nginx/sites-enabled/" + host)
			if err != nil {
				log.Fatal(err)
			}
		}

		hosts, _ := ioutil.ReadFile("/etc/hosts")

		reg := regexp.MustCompile("\n#host-" + host + "-begin\n127.0.0.1    " + host + "\n#host-" + host + "-end")
		res := reg.ReplaceAllString(string(hosts), "${1}")

		if err := ioutil.WriteFile("/etc/hosts", []byte(res), 0); err != nil {
			panic(err)
		}

		err := os.Remove("/etc/nginx/sites-available/" + host)
		if err != nil {
			log.Fatal(err)
		}
	default:
		fmt.Println("Unknown command. Hint: nxutil (create|enable|disable|remove) [HOSTNAME]")
	}
}
