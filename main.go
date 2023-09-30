package main

import (
        "fmt"
        "io/ioutil"
        "net/http"
        "os/exec"
        "regexp"
)

func getIP(w http.ResponseWriter, r *http.Request) {
        ip := r.Header.Get("X-Real-IP")
        if ip == "" {
                ip = r.RemoteAddr
        }

        err := addToWhiteList(ip)
        if err != nil {
                fmt.Println("Error:", err)
                // Handle the error response as needed.
                http.Error(w, "Internal Server Error", http.StatusInternalServerError)
                return
        }
}

func addToWhiteList(newIPAddress string) error {
        serviceFilePath := "/etc/systemd/system/caddy.service"

        content, err := ioutil.ReadFile(serviceFilePath)
        if err != nil {
                fmt.Println("Error reading file:", err)
                return err
        }

        originalString := string(content)

        re := regexp.MustCompile(`(WHITE_LIST=)([^"\s]+)`)

        match := re.FindString(originalString)

        if match != "" {
                existingWhiteList := re.FindStringSubmatch(originalString)[2]

                newWhiteList := existingWhiteList + " " + newIPAddress

                updatedString := re.ReplaceAllString(originalString, "${1}"+newWhiteList)

                err = ioutil.WriteFile(serviceFilePath, []byte(updatedString), 0644)
                if err != nil {
                        fmt.Println("Error writing file:", err)
                        return err
                }

                if err := exec.Command("systemctl", "daemon-reload").Run(); err != nil {
                        return err
                }

                if err := exec.Command("systemctl", "restart", "caddy.service").Run(); err != nil {
                        return err
                }

                fmt.Println("IP successfully bleached - ", newIPAddress)

                return nil
        } else {
                fmt.Println("WHITE_LIST not found.")
                return nil
        }
}

func main() {
        http.HandleFunc("/", getIP)

        port := 50009
        fmt.Printf("Server is running on port %d...\n", port)

        err := http.ListenAndServe(fmt.Sprintf(":%d", port), nil)
        if err != nil {
                fmt.Println("Error starting server:", err)
        }
}
