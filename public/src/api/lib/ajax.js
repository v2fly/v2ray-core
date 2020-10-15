import axios from 'axios';

function getBaseURL() {
    let baseURL = localStorage.getItem("baseURL") || "http://localhost:8035/v2ray";
    return baseURL;
}

let apiAuth = localStorage.getItem("api_auth");
if(apiAuth) {
    try{
        axios.defaults.auth = JSON.parse(apiAuth);
    }catch (e) {
        console.log("解析apiAuth失败：", e.toString());
    }
}
axios.defaults.baseURL = getBaseURL();
// axios.defaults.auth  = {
//     username: 'admin',
//     password: 'admin'
// };
// 添加响应拦截器
axios.interceptors.response.use(function (response) {
    // 对响应数据做点什么
    return response;
}, function (error) {
    if(error.response && error.response.status=="401"){
        ajax.openLoginForm();
    }
    return Promise.reject(error);
});

class Ajax {
    constructor(){

    }
    async post(url, data, options = {}) {
        try {
            return await axios.post(url, data, options);
        } catch (e) {
            return {
                status: 500,
                statusText: e.toString(),
            }
        }

    }
    async get(url, options) {
        try {
            return await axios.get(url, options);
        } catch (e) {
            return {
                status: 500,
                statusText: e.toString(),
            }
        }
    }
    setBaseURL(baseURL) {
        axios.defaults.baseURL = baseURL;
        localStorage.setItem("baseURL", baseURL);
    }
    getBaseURL(){
        return getBaseURL();
    }
    openLoginForm() {
    }
    setAuth(userName, password) {
        if(userName){
            axios.defaults.auth = {
                username: userName,
                password: password
            }
            localStorage.setItem("api_auth", JSON.stringify(axios.defaults.auth));
        }else{
            delete axios.defaults.auth;
            localStorage.removeItem("api_auth");
        }

    }
}
const ajax = ((function(){
    let ajax= null;
    return function(){
        if(ajax==null){
            ajax = new Ajax();
        }
        return ajax;
    }
})()());

export default ajax;
