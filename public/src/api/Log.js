import ajax from '@/api/lib/ajax'

export default {
    async loadLogContent({logType="access", from=-1}) {
        let res = {
            bSuccess: true,
            msg:""
        };
        let httpRes = await ajax.get(`/api/log?logType=${logType}&from=${from}`);
        if(httpRes.status!==200) {
            res.bSuccess = false;
            res.msg = "http调用失败:"+httpRes.statusText;
        }else{
            res.bSuccess = true;
            res.data = httpRes.data;
        }
        return res;
    },

}
