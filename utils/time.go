package utils

import (
  "time"
)


func RunForever(f func(), d time.Duration) {
  for {
    time.Sleep(d)
    f()
  }
}
