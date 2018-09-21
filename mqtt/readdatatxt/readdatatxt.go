package readdatatxt

import (   
    "bufio"    
    "io"
    "os"
    "strings"
)



type Config map[string]string

func ReadConfig(filename string) (Config, error) {
    // init with some bogus data
    config := Config{
        "nwsHexKey":     "padraoa8773564ebc8f7abdcaac6bd2137dd07",
        "appHexKey":     "padrao1e745697990164f22531bf11e4614ad1",
        "devHexAddr":    "padrao018a6355",
	"broker":        "padraotcp://localhost:1884",
	"username":       "padrao",
	"password":       "padrao",	

    }
    if len(filename) == 0 {
        return config, nil
    }
    file, err := os.Open(filename)
    if err != nil {
        return nil, err
    }
    defer file.Close()

    reader := bufio.NewReader(file)

    for {
        line, err := reader.ReadString('\n')

        // check if the line has = sign
        // and process the line. Ignore the rest.
        if equal := strings.Index(line, "="); equal >= 0 {
            if key := strings.TrimSpace(line[:equal]); len(key) > 0 {
                value := ""
                if len(line) > equal {
                    value = strings.TrimSpace(line[equal+1:])
                }
                // assign the config map
                config[key] = value
            }
        }
        if err == io.EOF {
            break
        }
        if err != nil {
            return nil, err
        }
    }
    return config, nil
}
