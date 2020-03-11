/*
	Copyright 2019 Awn Umar <awn@spacetime.dev>

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

package stdin

import (
	"errors"
	"os"

	"github.com/awnumar/memguard"
)

// ReadKeyFromStdin reads a key from standard inputs and returns it sealed inside an Enclave object.
func ReadKeyFromStdin() (*memguard.Enclave, error) {
	key, err := memguard.NewBufferFromReaderUntil(os.Stdin, '\n')
	if err != nil {
		// error encountered before '\n' was reached
		return nil, err
	}
	if key.Size() == 0 {
		return nil, errors.New("no input received")
	}
	return key.Seal(), nil
}
