package resource

import (
  "strconv"
  "strings"
)

// parsePrice extracts numeric value from price string
func parsePrice(priceStr string) float64 {
  parts := strings.Fields(priceStr)
  if len(parts) > 0 {
    val, _ := strconv.ParseFloat(parts[0], 64)
    return val
  }
  return 0.0
}

// truncateString truncates string to max length
func truncateString(s string, maxLen int) string {
  if len(s) <= maxLen {
    return s
  }
  return s[:maxLen-3] + "..."
}

// extractVCPUCount extracts numeric vCPU value for sorting
func extractVCPUCount(vCPUStr string) int {
  val, _ := strconv.Atoi(strings.TrimSpace(vCPUStr))
  return val
}

// extractRAMSize extracts numeric RAM value in GiB for sorting
func extractRAMSize(ramStr string) float64 {
  parts := strings.Fields(ramStr)
  if len(parts) > 0 {
    if val, err := strconv.ParseFloat(parts[0], 64); err == nil {
      return val
    }
  }
  return 0.0
}
