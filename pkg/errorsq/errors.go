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

package errorsq

import "context"

type Errorsq interface {
	Send(error)
}

type errorq struct {
	ctx    context.Context
	errCh  chan error
	logger func(error)
}

func NewErrorsq(ctx context.Context, logger func(error)) Errorsq {
	errs := &errorq{
		ctx:    ctx,
		logger: logger,
		errCh:  make(chan error, 10),
	}
	go func() {
		for {
			select {
			case <-errs.ctx.Done():
				close(errs.errCh)
				return
			case err, ok := <-errs.errCh:
				if !ok {
					return
				}
				errs.logger(err)
			}
		}
	}()
	return errs
}

func (errs *errorq) Send(err error) {
	errs.errCh <- err
}
