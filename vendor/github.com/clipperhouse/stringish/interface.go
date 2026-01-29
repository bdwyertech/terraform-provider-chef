// Copyright (c) HashiCorp, Inc.

package stringish

type Interface interface {
	~[]byte | ~string
}
