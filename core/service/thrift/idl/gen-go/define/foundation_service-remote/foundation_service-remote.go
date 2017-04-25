// Autogenerated by Thrift Compiler (0.9.3)
// DO NOT EDIT UNLESS YOU ARE SURE THAT YOU KNOW WHAT YOU ARE DOING

package main

import (
	"define"
	"flag"
	"fmt"
	"git.apache.org/thrift.git/lib/go/thrift"
	"math"
	"net"
	"net/url"
	"os"
	"strconv"
	"strings"
)

func Usage() {
	fmt.Fprintln(os.Stderr, "Usage of ", os.Args[0], " [-h host:port] [-u url] [-f[ramed]] function [arg1 [arg2...]]:")
	flag.PrintDefaults()
	fmt.Fprintln(os.Stderr, "\nFunctions:")
	fmt.Fprintln(os.Stderr, "  string ResourceUrl(string url)")
	fmt.Fprintln(os.Stderr, "  PlatformConf GetPlatformConf()")
	fmt.Fprintln(os.Stderr, "  string GetValue(string key)")
	fmt.Fprintln(os.Stderr, "  Result SetValue(string key, string value)")
	fmt.Fprintln(os.Stderr, "  Result DeleteValue(string key)")
	fmt.Fprintln(os.Stderr, "   GetValuesByPrefix(string prefix)")
	fmt.Fprintln(os.Stderr, "  string RegisterApp(SsoApp app)")
	fmt.Fprintln(os.Stderr, "  SsoApp GetApp(string name)")
	fmt.Fprintln(os.Stderr, "   GetAllSsoApp()")
	fmt.Fprintln(os.Stderr, "  bool ValidateSuper(string user, string pwd)")
	fmt.Fprintln(os.Stderr, "  void FlushSuperPwd(string user, string pwd)")
	fmt.Fprintln(os.Stderr, "  string GetSyncLoginUrl(string returnUrl)")
	fmt.Fprintln(os.Stderr)
	os.Exit(0)
}

func main() {
	flag.Usage = Usage
	var host string
	var port int
	var protocol string
	var urlString string
	var framed bool
	var useHttp bool
	var parsedUrl url.URL
	var trans thrift.TTransport
	_ = strconv.Atoi
	_ = math.Abs
	flag.Usage = Usage
	flag.StringVar(&host, "h", "localhost", "Specify host and port")
	flag.IntVar(&port, "p", 9090, "Specify port")
	flag.StringVar(&protocol, "P", "binary", "Specify the protocol (binary, compact, simplejson, json)")
	flag.StringVar(&urlString, "u", "", "Specify the url")
	flag.BoolVar(&framed, "framed", false, "Use framed transport")
	flag.BoolVar(&useHttp, "http", false, "Use http")
	flag.Parse()

	if len(urlString) > 0 {
		parsedUrl, err := url.Parse(urlString)
		if err != nil {
			fmt.Fprintln(os.Stderr, "Error parsing URL: ", err)
			flag.Usage()
		}
		host = parsedUrl.Host
		useHttp = len(parsedUrl.Scheme) <= 0 || parsedUrl.Scheme == "http"
	} else if useHttp {
		_, err := url.Parse(fmt.Sprint("http://", host, ":", port))
		if err != nil {
			fmt.Fprintln(os.Stderr, "Error parsing URL: ", err)
			flag.Usage()
		}
	}

	cmd := flag.Arg(0)
	var err error
	if useHttp {
		trans, err = thrift.NewTHttpClient(parsedUrl.String())
	} else {
		portStr := fmt.Sprint(port)
		if strings.Contains(host, ":") {
			host, portStr, err = net.SplitHostPort(host)
			if err != nil {
				fmt.Fprintln(os.Stderr, "error with host:", err)
				os.Exit(1)
			}
		}
		trans, err = thrift.NewTSocket(net.JoinHostPort(host, portStr))
		if err != nil {
			fmt.Fprintln(os.Stderr, "error resolving address:", err)
			os.Exit(1)
		}
		if framed {
			trans = thrift.NewTFramedTransport(trans)
		}
	}
	if err != nil {
		fmt.Fprintln(os.Stderr, "Error creating transport", err)
		os.Exit(1)
	}
	defer trans.Close()
	var protocolFactory thrift.TProtocolFactory
	switch protocol {
	case "compact":
		protocolFactory = thrift.NewTCompactProtocolFactory()
		break
	case "simplejson":
		protocolFactory = thrift.NewTSimpleJSONProtocolFactory()
		break
	case "json":
		protocolFactory = thrift.NewTJSONProtocolFactory()
		break
	case "binary", "":
		protocolFactory = thrift.NewTBinaryProtocolFactoryDefault()
		break
	default:
		fmt.Fprintln(os.Stderr, "Invalid protocol specified: ", protocol)
		Usage()
		os.Exit(1)
	}
	client := define.NewFoundationServiceClientFactory(trans, protocolFactory)
	if err := trans.Open(); err != nil {
		fmt.Fprintln(os.Stderr, "Error opening socket to ", host, ":", port, " ", err)
		os.Exit(1)
	}

	switch cmd {
	case "ResourceUrl":
		if flag.NArg()-1 != 1 {
			fmt.Fprintln(os.Stderr, "ResourceUrl requires 1 args")
			flag.Usage()
		}
		argvalue0 := flag.Arg(1)
		value0 := argvalue0
		fmt.Print(client.ResourceUrl(value0))
		fmt.Print("\n")
		break
	case "GetPlatformConf":
		if flag.NArg()-1 != 0 {
			fmt.Fprintln(os.Stderr, "GetPlatformConf requires 0 args")
			flag.Usage()
		}
		fmt.Print(client.GetPlatformConf())
		fmt.Print("\n")
		break
	case "GetValue":
		if flag.NArg()-1 != 1 {
			fmt.Fprintln(os.Stderr, "GetValue requires 1 args")
			flag.Usage()
		}
		argvalue0 := flag.Arg(1)
		value0 := argvalue0
		fmt.Print(client.GetValue(value0))
		fmt.Print("\n")
		break
	case "SetValue":
		if flag.NArg()-1 != 2 {
			fmt.Fprintln(os.Stderr, "SetValue requires 2 args")
			flag.Usage()
		}
		argvalue0 := flag.Arg(1)
		value0 := argvalue0
		argvalue1 := flag.Arg(2)
		value1 := argvalue1
		fmt.Print(client.SetValue(value0, value1))
		fmt.Print("\n")
		break
	case "DeleteValue":
		if flag.NArg()-1 != 1 {
			fmt.Fprintln(os.Stderr, "DeleteValue requires 1 args")
			flag.Usage()
		}
		argvalue0 := flag.Arg(1)
		value0 := argvalue0
		fmt.Print(client.DeleteValue(value0))
		fmt.Print("\n")
		break
	case "GetValuesByPrefix":
		if flag.NArg()-1 != 1 {
			fmt.Fprintln(os.Stderr, "GetValuesByPrefix requires 1 args")
			flag.Usage()
		}
		argvalue0 := flag.Arg(1)
		value0 := argvalue0
		fmt.Print(client.GetValuesByPrefix(value0))
		fmt.Print("\n")
		break
	case "RegisterApp":
		if flag.NArg()-1 != 1 {
			fmt.Fprintln(os.Stderr, "RegisterApp requires 1 args")
			flag.Usage()
		}
		arg44 := flag.Arg(1)
		mbTrans45 := thrift.NewTMemoryBufferLen(len(arg44))
		defer mbTrans45.Close()
		_, err46 := mbTrans45.WriteString(arg44)
		if err46 != nil {
			Usage()
			return
		}
		factory47 := thrift.NewTSimpleJSONProtocolFactory()
		jsProt48 := factory47.GetProtocol(mbTrans45)
		argvalue0 := define.NewSsoApp()
		err49 := argvalue0.Read(jsProt48)
		if err49 != nil {
			Usage()
			return
		}
		value0 := argvalue0
		fmt.Print(client.RegisterApp(value0))
		fmt.Print("\n")
		break
	case "GetApp":
		if flag.NArg()-1 != 1 {
			fmt.Fprintln(os.Stderr, "GetApp requires 1 args")
			flag.Usage()
		}
		argvalue0 := flag.Arg(1)
		value0 := argvalue0
		fmt.Print(client.GetApp(value0))
		fmt.Print("\n")
		break
	case "GetAllSsoApp":
		if flag.NArg()-1 != 0 {
			fmt.Fprintln(os.Stderr, "GetAllSsoApp requires 0 args")
			flag.Usage()
		}
		fmt.Print(client.GetAllSsoApp())
		fmt.Print("\n")
		break
	case "ValidateSuper":
		if flag.NArg()-1 != 2 {
			fmt.Fprintln(os.Stderr, "ValidateSuper requires 2 args")
			flag.Usage()
		}
		argvalue0 := flag.Arg(1)
		value0 := argvalue0
		argvalue1 := flag.Arg(2)
		value1 := argvalue1
		fmt.Print(client.ValidateSuper(value0, value1))
		fmt.Print("\n")
		break
	case "FlushSuperPwd":
		if flag.NArg()-1 != 2 {
			fmt.Fprintln(os.Stderr, "FlushSuperPwd requires 2 args")
			flag.Usage()
		}
		argvalue0 := flag.Arg(1)
		value0 := argvalue0
		argvalue1 := flag.Arg(2)
		value1 := argvalue1
		fmt.Print(client.FlushSuperPwd(value0, value1))
		fmt.Print("\n")
		break
	case "GetSyncLoginUrl":
		if flag.NArg()-1 != 1 {
			fmt.Fprintln(os.Stderr, "GetSyncLoginUrl requires 1 args")
			flag.Usage()
		}
		argvalue0 := flag.Arg(1)
		value0 := argvalue0
		fmt.Print(client.GetSyncLoginUrl(value0))
		fmt.Print("\n")
		break
	case "":
		Usage()
		break
	default:
		fmt.Fprintln(os.Stderr, "Invalid function ", cmd)
	}
}
