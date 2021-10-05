/*
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 * http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
*/
// This code is based on https://github.com/hyperledger-labs/fabric-operations-console/blob/main/packages/stitch/src/libs/misc.ts

import { logger } from '../logger';

// exports
export { underscores_2_camelCase, pp };


// --------------------------------------------------------------------------------
// pretty print JSON - truncates long objects!
// --------------------------------------------------------------------------------
// eslint-disable-next-line @typescript-eslint/ban-types
function pp(value: object | undefined | null) {
  try {
    let temp = JSON.stringify(value, null, '\t');
    if (temp && temp.length >= 25000) {
      temp = temp.substring(0, 25000) + '... [too long, truncated the rest]';
    }
    return temp;
  } catch (e) {
    return value;
  }
}

// --------------------------------------------------------------------------------
// change key's in an object from using underscores to camel case (existing case is preserved where applicable)
// --------------------------------------------------------------------------------
function underscores_2_camelCase(orig: any, _iter: number | null): any {
  if (!_iter) { _iter = 1; }
  if (typeof orig !== 'object') {
    logger.warn('[protobuf-handler] underscores_2_camelCase() is expecting an object. not:', orig);
    return null;
  } else if (_iter >= 1000) {
    logger.error('[protobuf-handler] too many recursive loops, cannot convert obj:', orig, _iter);
    return null;
  } else {
    const ret: any = {};

    if (Array.isArray(orig)) {				// if its an array, see if array contains objects
      const arr = [];
      for (const i in orig) {															// iter on array contents
        if (typeof orig[i] === 'object') {
          arr.push(underscores_2_camelCase(orig[i], ++_iter));					// recursive
        } else {
          arr.push(orig[i]);
        }
      }
      return arr;
    } else {
      for (const key in orig) {
        const parts = key.split('_');
        const formatted = [];														// ts won't let me overwrite parts
        for (const i in parts) {
          if (Number(i) === 0) {
            formatted.push(parts[i]);											// first word is already good
          } else {
            formatted.push(parts[i][0].toUpperCase() + parts[i].substring(1));	// convert first letter to uppercase
          }
        }

        if (formatted.length === 0) {
          logger.warn('[protobuf-handler] underscores_2_camelCase() cannot format key:', parts);
        } else {
          const newKey = formatted.join('');
          if (typeof orig[key] === 'object') {
            ret[newKey] = underscores_2_camelCase(orig[key], ++_iter);			// recursive
          } else {
            ret[newKey] = orig[key];
          }
        }
      }
    }
    return ret;
  }
}