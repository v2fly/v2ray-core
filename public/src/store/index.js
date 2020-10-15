import Vue from 'vue'
import Vuex from 'vuex'

Vue.use(Vuex)
function removeRepeat(arr) {
    let newArr = [];
    let keyMap = {}
    arr.forEach(a=>{
        if(typeof(keyMap[a])!=="undefined"){
            return;
        }
        keyMap[a] = a;
        newArr.push(a);
    });
    return newArr;
}
export default new Vuex.Store({
    state: {
        inboundTags: [],
        bridgeTags: [],
        portalTags: [],
        outboundTags: [],
        apiTag: null,
        balancersTag:[],
        dnsTag: null,
    },
    mutations: {
        setInboundTags(state, inboundTags) {
            state.inboundTags = inboundTags;
        },
        setDnsTag(state, dnsTag) {
            state.dnsTag = dnsTag;
        },
        setOutboundTags(state, outboundTags) {
            state.outboundTags = outboundTags;
        },
        setBridgeTags(state, bridgeTags) {
            state.bridgeTags = bridgeTags;
        },
        setPortalTags(state, portalTags) {
            state.portalTags = portalTags;
        },
        setApiTag(state, tag) {
            state.apiTag = tag;
        },
        setBalancersTag(state, balancersTag) {
            state.balancersTag = balancersTag;
        },
    },
    actions: {},
    getters: {
        getAllInboundTags(state) {
            let tags = [].concat(...state.inboundTags);
            if(state.bridgeTags){
                tags.push(...state.bridgeTags);
            }
            if(state.dnsTag) {
                tags.push(state.dnsTag);
            }
            return removeRepeat(tags);
        },
        getAllOutboundTags(state) {
            let tags = [];
            if (state.apiTag) {
                tags.push(state.apiTag);
            }
            tags.push(...state.outboundTags);
            if(state.portalTags){
                tags.push(...state.portalTags);
            }
            return removeRepeat(tags);
        },
        getBalancersTag(state) {
            return removeRepeat(state.balancersTag);
        }
    },
    modules: {}
})
