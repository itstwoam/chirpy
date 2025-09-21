package config

import ("os"
	"fmt"
	"errors"
	"encoding/json"
)

type Config struct {
	DB_url	string	`json:"db_url"`
	Default_Expiry int `json:"default_expiry"`
}

const configFileName = ".chirpy.json"


func Read() (*Config, error){
	filePath,err := getConfigFilePath()
	if err != nil{
		fmt.Println("in Read, error from getConfigFilePath()")
		return nil, err
	}
	fmt.Println("filePath: ",filePath)
	file, err := os.Open(filePath)
	if err != nil{
		fmt.Println("in Read, error from os.Open")
		fmt.Println(err)
		return nil, err
	}
	defer file.Close()
	var final Config
	if err := json.NewDecoder(file).Decode(&final); err != nil {
		fmt.Println("in Read, error from decode")
		return nil, err
	}
	return &final, nil
}

func getConfigFilePath() (string, error){
	homeDir, err := os.UserHomeDir()
	if err != nil{
		fmt.Println("in getConfigFilePath(), error from os.UserHomeDir()")
		return "",errors.New("home directory not found")
	}
	homeDir = homeDir + "/"+ configFileName
	return homeDir, nil
}

func write(cfg *Config) error {
	filepath, err := getConfigFilePath()
	if err != nil {
		return errors.New("unable to find path to write")
	}
	file, err := os.Create(filepath)
	if err != nil {
		return err
	}
	defer file.Close()

	encoder := json.NewEncoder(file)
	encoder.SetIndent("", "  ")
	return encoder.Encode(cfg) 
}
