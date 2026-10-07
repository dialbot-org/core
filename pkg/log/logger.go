package logger

import (
	"fmt"
	"log"
)

const (
	Reset = "\033[0m"
	RedBright    = "\033[91m"
	GreenBright  = "\033[92m"
	YellowBright = "\033[93m"
	BlueBright   = "\033[94m"
	MagentaBright= "\033[95m"
	CyanBright   = "\033[96m"
	WhiteBright  = "\033[97m"
)

func logMessage(color, label, symbol, format string, args ...any) {
	message := fmt.Sprintf(format, args...)
	log.Printf("%s%s%s %s %s",
		color, symbol, label, message, Reset,
	)
}

func Info(format string, args ...any) {
	logMessage(CyanBright, "INFO", "ℹ️", format, args...)
}

func Success(format string, args ...any) {
	logMessage(GreenBright, "SUCCESS", "✅", format, args...)
}

func Warning(format string, args ...any) {
	logMessage(YellowBright, "WARNING", "⚠️", format, args...)
}

func Error(format string, args ...any) {
	logMessage(RedBright, "ERROR", "❌", format, args...)
}

func Debug(format string, args ...any) {
	logMessage(BlueBright, "DEBUG", "🦶🏻", format, args...)
}

func Fatal(err error, msg string) {
	if err != nil {
		log.Fatalf("%s☠ FATAL [%s: %s]%s",
			RedBright, msg, err, Reset,
		)
	}
}

func Start() {
	log.Printf(
		`%s
#################################################################

                                               @@@@@@@@         
                                          @@@@@@@@@@@@@@    	
                                     @@@@@@@@@@@@@@@@@@@    	
                                @@@@@@@@@@@@@@@@@@@@@@@@    	
                            @@@@@@@@@@@@@@@@@@@@@@@@@@@@    	
                       @@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@     	
                  @@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@     	
             @@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@      	
         @@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@               	
    @@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@                   	
 @@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@          @@@@      	
@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@            @@@@@@@   	
@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@             @@@@@@@@  	
@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@              @@@@@@@@@ 	
 @@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@              @@@@@@@@@@	
     @@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@    @@@@@@@@@@@@@@@@@@@@	
                   @@@@@@@@@@@@@@@@@     @@@@@@@@@@@@@@@@@@@	
                         @@@@@@@@@@@@    @@@@@@@@@@@@@@@@@@ 	
                         @@@@@@@@@@@@     @@@@@@@@@@@@@@@@  	
                         @@@@@@@@@@@@@     @@@@@@@@@@@@@@   	
                          @@@@@@@@@@@@@@      @@@@@@@@      	
                          @@@@@@@@@@@@@@@@                  	
                          @@@@@@@@@@@@@@@@                  	
                          @@@@@@@@@@@@@@@@                  	
                          @@@@@@@@@@@@@@@                   	
                           @@@@@@@@@@@@@                    	
                           @@@@@@@@@@@@@                    	
                           @@@@@@@@@@@@                     	
                            @@@@@@@@@@                      	
                             @@@@@@@@                       	
																
##################### AutoCall server ####################
		%s`,
		CyanBright,
		Reset,
	)
}