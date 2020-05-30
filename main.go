package main

import (
	"flag"
	"os"
	"strings"

	jwt "github.com/dgrijalva/jwt-go"
	utils "github.com/flaviostutz/go-utils"
	"github.com/jinzhu/gorm"
	"github.com/sirupsen/logrus"
)

type options struct {
	logLevel string

	corsAllowedOrigins string
	jwtSigningMethod   string
	jwtSigningKeyFile  string
	jwtPublicKey       interface{}

	baseURL    string
	sqliteFile string
}

var (
	opt options
	db  *gorm.DB
)

func main() {
	logLevel := flag.String("loglevel", "debug", "debug, info, warning, error")

	corsAllowedOrigins0 := flag.String("cors-allowed-origins", "*", "Cors allowed origins for this server")
	jwtSigningMethod0 := flag.String("jwt-signing-method", "", "JWT signing method. required")
	jwtSigningKeyFile0 := flag.String("jwt-signing-key-file", "", "Key file used to sign tokens. Tokens may be later validated by thirdy parties by checking the signature with related public key when usign assymetric keys")
	baseURL0 := flag.String("base-url", "", "Base URL used as a prefix for 'Location' headers")

	flag.Parse()

	switch *logLevel {
	case "debug":
		logrus.SetLevel(logrus.DebugLevel)
		break
	case "warning":
		logrus.SetLevel(logrus.WarnLevel)
		break
	case "error":
		logrus.SetLevel(logrus.ErrorLevel)
		break
	default:
		logrus.SetLevel(logrus.InfoLevel)
	}

	opt = options{
		logLevel: *logLevel,

		corsAllowedOrigins: *corsAllowedOrigins0,
		jwtSigningMethod:   *jwtSigningMethod0,
		jwtSigningKeyFile:  *jwtSigningKeyFile0,
		baseURL:            *baseURL0,

		sqliteFile: "/demo.db",
	}

	sm := jwt.GetSigningMethod(opt.jwtSigningMethod)
	if sm == nil {
		logrus.Errorf("Unsupported JWT signing method %s", opt.jwtSigningMethod)
		os.Exit(1)
	}

	logrus.Infof("Loading JWT public key")

	logrus.Debugf("JWT signing method: %s", opt.jwtSigningMethod)
	if strings.HasPrefix(opt.jwtSigningMethod, "RS") || strings.HasPrefix(opt.jwtSigningMethod, "ES") || strings.HasPrefix(opt.jwtSigningMethod, "HS") {
		pubk, err := utils.ParseKeyFromPEM(opt.jwtSigningKeyFile, false)
		if err != nil {
			logrus.Errorf("Failed to parse PEM public key. err=%s", err)
			os.Exit(1)
		}
		opt.jwtPublicKey = pubk
	} else {
		logrus.Errorf("Unsupported signing method %s", opt.jwtSigningMethod)
		os.Exit(1)
	}
	logrus.Debugf("JWT key loaded")

	db0, err0 := initDB()
	if err0 != nil {
		logrus.Warnf("Couldn't init database. err=%s", err0)
		os.Exit(1)
	}
	db = db0
	defer db.Close()

	err := NewHTTPServer().Start()
	if err != nil {
		logrus.Warnf("Error starting server. err=%s", err)
		os.Exit(1)
	}
}
