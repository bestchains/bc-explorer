/*
Copyright 2023 The Bestchains Authors.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package viewer

import (
	"k8s.io/klog/v2"
)

type overviewLogger struct {
}

func NewOverviewLogger() Overview {
	klog.Infoln("use overview logger handler")
	return &overviewLogger{}
}

func (o *overviewLogger) Summary(network string) (SummaryResp, error) {
	klog.Infof("overviewLogger Summary with network %s\n", network)
	return SummaryResp{BlockNumber: 1, TxCount: 1}, nil
}

func (o *overviewLogger) QueryBySeg(from, interval, number int64, which, network string) ([]BySegResp, error) {
	klog.Infof("overviewLogger QueryBySeg")
	klog.Infof("from=%s, interval=%d, number=%d, which=%s, network=%s\n", from, interval, number, which, network)
	return []BySegResp{{Start: 0, End: 5, Count: 5}}, nil
}
