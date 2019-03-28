package tools

import (
	"encoding/json"
	"errors"
	"fmt"
	"github.com/go-yaml/yaml"
	"github.com/ihaiker/tenured-go-server/commons"
	"github.com/ihaiker/tenured-go-server/services"
	"github.com/ihaiker/tenured-go-server/services/console"
	"github.com/ihaiker/tenured-go-server/services/store"
	"github.com/spf13/cobra"
	"strings"
)

var serverConfig map[string]interface{}

func init() {
	serverConfig = map[string]interface{}{}
	serverConfig["store"] = store.NewStoreConfig()
	serverConfig["console"] = console.NewConsoleConfig()
}

var ConfigCmd = &cobra.Command{
	Use:     "config",
	Short:   "Tenured Config Tools",
	Long:    `Complete documentation is available at http://tenured.renzhen.la/configuration`,
	Version: "1.0.0",
	//Args:    cobra.MinimumNArgs(1),
	Example: `	tenured config -s store -f <path> -t #test store config
	tenured config -s store -o <path>#println store config to <path>`,
	RunE: func(cmd *cobra.Command, args []string) (err error) {
		defer func() {
			if e := recover(); e != nil {
				err = commons.Catch(e)
			}
		}()
		test, err := cmd.PersistentFlags().GetBool("test")
		commons.Painc(err)

		print, err := cmd.PersistentFlags().GetBool("print")
		commons.Painc(err)
		printJson, err := cmd.PersistentFlags().GetBool("json")
		commons.Painc(err)
		printYaml, err := cmd.PersistentFlags().GetBool("yaml")
		commons.Painc(err)

		out, err := cmd.PersistentFlags().GetString("out")
		commons.Painc(err)

		if !test && !print && out == "" {
			return cmd.Usage()
		}

		config, err := cmd.PersistentFlags().GetString("config")
		commons.Painc(err)
		server, err := cmd.PersistentFlags().GetStringArray("server")
		commons.Painc(err)

		if len(server) == 0 {
			server = make([]string, 0)
			for k := range serverConfig {
				server = append(server, k)
			}
		}

		if print {
			var bs []byte
			if printJson || (!printJson && !printYaml) {
				for _, v := range server {
					if cfg, has := serverConfig[v]; has {
						bs, err = json.MarshalIndent(cfg, "\t", "")
						commons.Painc(err)
						fmt.Println(string(bs))
					} else {
						return errors.New("not found server: " + v)
					}
				}
			}
			if printYaml {
				for _, v := range server {
					if cfg, has := serverConfig[v]; has {
						bs, err = yaml.Marshal(cfg)
						commons.Painc(err)
						fmt.Println(string(bs))
					} else {
						return errors.New("not found server: " + v)
					}
				}
			}
		} else if test {
			for _, v := range server {
				if cfg, has := serverConfig[v]; has {
					commons.Painc(services.LoadServerConfig(v, config, cfg))
				} else {
					return errors.New("not found server: " + v)
				}
			}
			fmt.Println("test success!")
		} else if out != "" {
			cfg, has := serverConfig[server[0]]
			if !has {
				return errors.New("not found server: " + server[0])
			}
			var bs []byte
			if !printJson && !printYaml {
				printJson = strings.HasSuffix(out, ".json")
				printYaml = strings.HasSuffix(out, ".yaml") || strings.HasSuffix(out, ".yml")
			}

			if printJson || (!printJson && !printYaml) {
				bs, err = json.MarshalIndent(cfg, "\t", "")
				commons.Painc(err)
			} else if printYaml {
				bs, err = yaml.Marshal(cfg)
				commons.Painc(err)
			}

			fs, err := commons.NewFile(out).GetWriter(false)
			commons.Painc(err)
			defer fs.Close()
			_, err = fs.Write(bs)
			commons.Painc(err)
			fmt.Println("write success!")
		}
		return nil
	},
}

func init() {
	ConfigCmd.PersistentFlags().StringP("config", "f", "",
		`the config file`)

	ConfigCmd.PersistentFlags().BoolP("test", "t", false,
		`test the config file`)

	ConfigCmd.PersistentFlags().StringArrayP("server", "s", []string{},
		`the server name. {store|linker|console|push|restful}`)

	ConfigCmd.PersistentFlags().StringP("out", "o", "",
		`out config file`)

	ConfigCmd.PersistentFlags().BoolP("print", "p", false,
		`print config file`)

	ConfigCmd.PersistentFlags().BoolP("json", "", false,
		`print json config demo`)

	ConfigCmd.PersistentFlags().BoolP("yaml", "", false,
		`print yaml config demo`)
}
