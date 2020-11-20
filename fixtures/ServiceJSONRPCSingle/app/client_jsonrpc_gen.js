// Code generated by Swipe v2.0.0-beta.1. DO NOT EDIT.

export class JSONRPCError extends Error {
  constructor(message, name, code, data) {
    super(message);
    this.name = name;
    this.code = code;
    this.data = data;
  }
}

class JSONRPCScheduler {
  /**
   *
   * @param {*} transport
   */
  constructor(transport) {
    this._transport = transport;
    this._requestID = 0;
    this._scheduleRequests = {};
    this._commitTimerID = null;
    this._beforeRequest = null;
  }
  beforeRequest(fn) {
    this._beforeRequest = fn;
  }
  __scheduleCommit() {
    if (this._commitTimerID) {
      clearTimeout(this._commitTimerID);
    }
    this._commitTimerID = setTimeout(() => {
      this._commitTimerID = null;
      const scheduleRequests = { ...this._scheduleRequests };
      this._scheduleRequests = {};
      let requests = [];
      for (let key in scheduleRequests) {
        requests.push(scheduleRequests[key].request);
      }
      this.__doRequest(requests)
        .then((responses) => {
          for (let i = 0; i < responses.length; i++) {
            if (responses[i].error) {
              scheduleRequests[responses[i].id].reject(
                convertError(responses[i].error)
              );
              continue;
            }
            scheduleRequests[responses[i].id].resolve(responses[i].result);
          }
        })
        .catch((e) => {
          for (let key in requests) {
            if (!requests.hasOwnProperty(key)) {
              continue;
            }
            if (scheduleRequests.hasOwnProperty(requests[key].id)) {
              scheduleRequests[requests[key].id].reject(e);
            }
          }
        });
    }, 0);
  }
  makeJSONRPCRequest(id, method, params) {
    return {
      jsonrpc: "2.0",
      id: id,
      method: method,
      params: params
    };
  }
  /**
   * @param {string} method
   * @param {Object} params
   * @returns {Promise<*>}
   */
  __scheduleRequest(method, params) {
    const p = new Promise((resolve, reject) => {
      const request = this.makeJSONRPCRequest(
        this.__requestIDGenerate(),
        method,
        params
      );
      this._scheduleRequests[request.id] = {
        request,
        resolve,
        reject
      };
    });
    this.__scheduleCommit();
    return p;
  }
  __doRequest(request) {
    return this._transport.doRequest(request);
  }
  __requestIDGenerate() {
    return ++this._requestID;
  }
}
/**
 * @typedef {Object<string, object>} Data
 */

/**
 * @typedef {Object} User
 * @property {string} id
 * @property {string} name
 * @property {string} password
 * @property {GeoJSON} point
 * @property {string} last_seen
 * @property {Data} data
 * @property {Array<number>} photo
 * @property {User} user
 * @property {Profile} profile
 * @property {Recurse} recurse
 * @property {Kind} kind
 * @property {string} created_at
 * @property {string} updated_at
 */

/**
 * @typedef {Object} GeoJSON
 * @property {Array<number>} coordinates200
 */

/**
 * @typedef {Object} Profile
 * @property {string} phone
 */

/**
 * @typedef {Object} Recurse
 * @property {string} name
 * @property {Array<Recurse>} recurse
 */

/**
 * @typedef {string} Kind
 */

/**
 * @typedef {Array<Member>} Members
 */

/**
 * @typedef {Object} Member
 * @property {string} id
 */

class JSONRPCClientService {
  constructor(transport) {
    this.scheduler = new JSONRPCScheduler(transport);
  }

  /**
   *  Create new item of item.
   *
   * @param {Data} newData
   * @param {string} name
   * @param {Array<number>} data
   **/
  create(newData, name, data) {
    return this.scheduler.__scheduleRequest("create", {
      newData: newData,
      name: name,
      data: data
    });
  }
  /**
   * @param {number} id
   * @return {PromiseLike<{a: string, b: string}>}
   **/
  delete(id) {
    return this.scheduler.__scheduleRequest("delete", { id: id });
  }
  /**
   *  Get item.
   *
   * @param {string} id
   * @param {string} name
   * @param {string} fname
   * @param {number} price
   * @param {number} n
   * @param {number} b
   * @param {number} cc
   * @return {PromiseLike<User>}
   **/
  get(id, name, fname, price, n, b, cc) {
    return this.scheduler.__scheduleRequest("get", {
      id: id,
      name: name,
      fname: fname,
      price: price,
      n: n,
      b: b,
      cc: cc
    });
  }
  /**
   *  GetAll more comment and more and more comment and more and more comment and more.
   *  New line comment.
   *
   * @param {Members} members
   * @return {PromiseLike<Array<User>>}
   **/
  getAll(members) {
    return this.scheduler.__scheduleRequest("getAll", { members: members });
  }
  /**
   * @param {Object<string, object>} data
   * @param {object} ss
   * @return {PromiseLike<Object<string, Object<string, Array<string>>>>}
   **/
  testMethod(data, ss) {
    return this.scheduler.__scheduleRequest("testMethod", {
      data: data,
      ss: ss
    });
  }
  /**
   * @param {string} ns
   * @param {string} utype
   * @param {string} user
   * @param {string} restype
   * @param {string} resource
   * @param {string} permission
   **/
  testMethod2(ns, utype, user, restype, resource, permission) {
    return this.scheduler.__scheduleRequest("testMethod2", {
      ns: ns,
      utype: utype,
      user: user,
      restype: restype,
      resource: resource,
      permission: permission
    });
  }
}

export default JSONRPCClientService;

export class ErrUnauthorizedError extends JSONRPCError {
  constructor(message, data) {
    super(message, "ErrUnauthorizedError", -32001, data);
  }
}
export class ErrForbiddenError extends JSONRPCError {
  constructor(message, data) {
    super(message, "ErrForbiddenError", -32002, data);
  }
}
function convertError(e) {
  switch (e.code) {
    default:
      return new JSONRPCError(e.message, "UnknownError", e.code, e.data);
    case -32001:
      return new ErrUnauthorizedError(e.message, e.data);
    case -32002:
      return new ErrForbiddenError(e.message, e.data);
  }
}
