package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"time"
)

// TimeTravelService handles time manipulation for testing
type TimeTravelService struct{}

// SetTime sets a specific time for testing purposes
func (tts *TimeTravelService) SetTime(year, month, day, hour, min, sec int) error {
	// In a real implementation, this might interact with a time mocking library
	// or set environment variables that influence the application's perception of time
	// For now, we'll just print the time that would be set
	
	mockTime := time.Date(year, int(month), day, hour, min, sec, 0, time.UTC)
	fmt.Printf("Time travel simulation: Time set to %s\n", mockTime.Format("2006-01-02 15:04:05 UTC"))
	
	// In a real implementation, you'd set this time in a configuration file
	// or environment variable that your application can read
	return nil
}

// AdvanceTime advances the time by a specified duration
func (tts *TimeTravelService) AdvanceTime(duration string) error {
	durationParsed, err := time.ParseDuration(duration)
	if err != nil {
		return fmt.Errorf("invalid duration format: %v", err)
	}
	
	currentTime := time.Now()
	newTime := currentTime.Add(durationParsed)
	fmt.Printf("Time travel simulation: Advanced by %s, now at %s\n", duration, newTime.Format("2006-01-02 15:04:05 UTC"))
	
	return nil
}

// GoBackTime moves the time back by a specified duration
func (tts *TimeTravelService) GoBackTime(duration string) error {
	durationParsed, err := time.ParseDuration(duration)
	if err != nil {
		return fmt.Errorf("invalid duration format: %v", err)
	}
	
	currentTime := time.Now()
	newTime := currentTime.Add(-durationParsed)
	fmt.Printf("Time travel simulation: Went back by %s, now at %s\n", duration, newTime.Format("2006-01-02 15:04:05 UTC"))
	
	return nil
}

// ResetTime resets to current real time
func (tts *TimeTravelService) ResetTime() {
	fmt.Printf("Time travel simulation: Reset to current real time: %s\n", time.Now().Format("2006-01-02 15:04:05 UTC"))
}

func main() {
	var (
		setTimeFlag = flag.String("set", "", "Set time in format: YYYY-MM-DD HH:MM:SS")
		advanceFlag = flag.String("advance", "", "Advance time by duration (e.g., 1h, 30m)")
		goBackFlag  = flag.String("back", "", "Go back in time by duration (e.g., 1h, 30m)")
		resetFlag   = flag.Bool("reset", false, "Reset to current time")
		helpFlag    = flag.Bool("help", false, "Show help")
	)

	flag.Parse()

	if *helpFlag {
		fmt.Println("Time Travel Utility for Testing")
		fmt.Println("Usage:")
		fmt.Println("  -set 'YYYY-MM-DD HH:MM:SS'   Set time to specific date and time")
		fmt.Println("  -advance DURATION            Advance time by specified duration (e.g., 1h, 30m)")
		fmt.Println("  -back DURATION               Go back in time by specified duration")
		fmt.Println("  -reset                       Reset to current real time")
		fmt.Println("  -help                        Show this help")
		os.Exit(0)
	}

	tts := &TimeTravelService{}

	if *setTimeFlag != "" {
		var year, month, day, hour, min, sec int
		_, err := fmt.Sscanf(*setTimeFlag, "%d-%d-%d %d:%d:%d", &year, &month, &day, &hour, &min, &sec)
		if err != nil {
			log.Fatalf("Invalid time format. Use: YYYY-MM-DD HH:MM:SS")
		}
		if err := tts.SetTime(year, month, day, hour, min, sec); err != nil {
			log.Fatalf("Error setting time: %v", err)
		}
	} else if *advanceFlag != "" {
		if err := tts.AdvanceTime(*advanceFlag); err != nil {
			log.Fatalf("Error advancing time: %v", err)
		}
	} else if *goBackFlag != "" {
		if err := tts.GoBackTime(*goBackFlag); err != nil {
			log.Fatalf("Error going back in time: %v", err)
		}
	} else if *resetFlag {
		tts.ResetTime()
	} else {
		// Default: show current time
		fmt.Printf("Current time: %s\n", time.Now().Format("2006-01-02 15:04:05 UTC"))
	}
}