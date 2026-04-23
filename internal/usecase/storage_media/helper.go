package storage_media_usecase

type byteRange struct {
	start int64
	end   int64
}

// func parseRangeHeader(rangeHeader string, totalSize int64) (byteRange, bool, error) {
// 	rangeHeader = strings.TrimSpace(rangeHeader)
// 	if rangeHeader == "" {
// 		return byteRange{}, false, nil
// 	}
// 	if totalSize <= 0 {
// 		return byteRange{}, false, errs.ErrGenericRangeNotSatisfiable
// 	}
// 	lowerHeader := strings.ToLower(rangeHeader)
// 	rangeSpec := rangeHeader
// 	if strings.HasPrefix(lowerHeader, "bytes=") {
// 		rangeSpec = strings.TrimSpace(rangeHeader[len("bytes="):])
// 	}
// 	if idx := strings.Index(rangeSpec, ","); idx >= 0 {
// 		rangeSpec = rangeSpec[:idx]
// 	}
// 	rangeSpec = strings.TrimSpace(rangeSpec)
// 	parts := strings.SplitN(rangeSpec, "-", 2)
// 	if len(parts) != 2 {
// 		return byteRange{}, false, errs.ErrGenericRangeNotSatisfiable
// 	}
// 	parts[0] = strings.TrimSpace(parts[0])
// 	parts[1] = strings.TrimSpace(parts[1])
// 	if parts[0] == "" {
// 		suffixLength, err := strconv.ParseInt(parts[1], 10, 64)
// 		if err != nil || suffixLength <= 0 {
// 			return byteRange{}, false, errs.ErrGenericRangeNotSatisfiable
// 		}
// 		if suffixLength > totalSize {
// 			suffixLength = totalSize
// 		}
// 		return byteRange{start: totalSize - suffixLength, end: totalSize - 1}, true, nil
// 	}
// 	start, err := strconv.ParseInt(parts[0], 10, 64)
// 	if err != nil || start < 0 || start >= totalSize {
// 		return byteRange{}, false, errs.ErrGenericRangeNotSatisfiable
// 	}
// 	var end int64
// 	if parts[1] == "" {
// 		end = totalSize - 1
// 	} else {
// 		end, err = strconv.ParseInt(parts[1], 10, 64)
// 		if err != nil || end < start {
// 			return byteRange{}, false, errs.ErrGenericRangeNotSatisfiable
// 		}
// 		if end >= totalSize {
// 			end = totalSize - 1
// 		}
// 	}
// 	return byteRange{start: start, end: end}, true, nil
// }
