import { txClient, queryClient, MissingWalletError , registry} from './module'

import { Query } from "./module/types/interchainquery/genesis"
import { DataPoint } from "./module/types/interchainquery/genesis"


export { Query, DataPoint };

async function initTxClient(vuexGetters) {
	return await txClient(vuexGetters['common/wallet/signer'], {
		addr: vuexGetters['common/env/apiTendermint']
	})
}

async function initQueryClient(vuexGetters) {
	return await queryClient({
		addr: vuexGetters['common/env/apiCosmos']
	})
}

function mergeResults(value, next_values) {
	for (let prop of Object.keys(next_values)) {
		if (Array.isArray(next_values[prop])) {
			value[prop]=[...value[prop], ...next_values[prop]]
		}else{
			value[prop]=next_values[prop]
		}
	}
	return value
}

function getStructure(template) {
	let structure = { fields: [] }
	for (const [key, value] of Object.entries(template)) {
		let field: any = {}
		field.name = key
		field.type = typeof value
		structure.fields.push(field)
	}
	return structure
}

const getDefaultState = () => {
	return {
				SubmitQueryResponse: {},
				
				_Structure: {
						Query: getStructure(Query.fromPartial({})),
						DataPoint: getStructure(DataPoint.fromPartial({})),
						
		},
		_Registry: registry,
		_Subscriptions: new Set(),
	}
}

// initial state
const state = getDefaultState()

export default {
	namespaced: true,
	state,
	mutations: {
		RESET_STATE(state) {
			Object.assign(state, getDefaultState())
		},
		QUERY(state, { query, key, value }) {
			state[query][JSON.stringify(key)] = value
		},
		SUBSCRIBE(state, subscription) {
			state._Subscriptions.add(JSON.stringify(subscription))
		},
		UNSUBSCRIBE(state, subscription) {
			state._Subscriptions.delete(JSON.stringify(subscription))
		}
	},
	getters: {
				getSubmitQueryResponse: (state) => (params = { params: {}}) => {
					if (!(<any> params).query) {
						(<any> params).query=null
					}
			return state.SubmitQueryResponse[JSON.stringify(params)] ?? {}
		},
				
		getTypeStructure: (state) => (type) => {
			return state._Structure[type].fields
		},
		getRegistry: (state) => {
			return state._Registry
		}
	},
	actions: {
		init({ dispatch, rootGetters }) {
			console.log('Vuex module: stride.interchainquery initialized!')
			if (rootGetters['common/env/client']) {
				rootGetters['common/env/client'].on('newblock', () => {
					dispatch('StoreUpdate')
				})
			}
		},
		resetState({ commit }) {
			commit('RESET_STATE')
		},
		unsubscribe({ commit }, subscription) {
			commit('UNSUBSCRIBE', subscription)
		},
		async StoreUpdate({ state, dispatch }) {
			state._Subscriptions.forEach(async (subscription) => {
				try {
					const sub=JSON.parse(subscription)
					await dispatch(sub.action, sub.payload)
				}catch(e) {
					throw new Error('Subscriptions: ' + e.message)
				}
			})
		},
		
		
		
		 		
		
		
		async MsgSubmitQueryResponse({ commit, rootGetters, getters }, { options: { subscribe, all} = { subscribe:false, all:false}, params, query=null }) {
			try {
				const key = params ?? {};
				const queryClient=await initQueryClient(rootGetters)
				let value= (await queryClient.msgSubmitQueryResponse({...key})).data
				
					
				commit('QUERY', { query: 'SubmitQueryResponse', key: { params: {...key}, query}, value })
				if (subscribe) commit('SUBSCRIBE', { action: 'MsgSubmitQueryResponse', payload: { options: { all }, params: {...key},query }})
				return getters['getSubmitQueryResponse']( { params: {...key}, query}) ?? {}
			} catch (e) {
				throw new Error('QueryClient:MsgSubmitQueryResponse API Node Unavailable. Could not perform query: ' + e.message)
				
			}
		},
		
		
		async sendMsgQueryBalance({ rootGetters }, { value, fee = [], memo = '' }) {
			try {
				const txClient=await initTxClient(rootGetters)
				const msg = await txClient.msgQueryBalance(value)
				const result = await txClient.signAndBroadcast([msg], {fee: { amount: fee, 
	gas: "200000" }, memo})
				return result
			} catch (e) {
				if (e == MissingWalletError) {
					throw new Error('TxClient:MsgQueryBalance:Init Could not initialize signing client. Wallet is required.')
				}else{
					throw new Error('TxClient:MsgQueryBalance:Send Could not broadcast Tx: '+ e.message)
				}
			}
		},
		async sendMsgSubmitTx({ rootGetters }, { value, fee = [], memo = '' }) {
			try {
				const txClient=await initTxClient(rootGetters)
				const msg = await txClient.msgSubmitTx(value)
				const result = await txClient.signAndBroadcast([msg], {fee: { amount: fee, 
	gas: "200000" }, memo})
				return result
			} catch (e) {
				if (e == MissingWalletError) {
					throw new Error('TxClient:MsgSubmitTx:Init Could not initialize signing client. Wallet is required.')
				}else{
					throw new Error('TxClient:MsgSubmitTx:Send Could not broadcast Tx: '+ e.message)
				}
			}
		},
		async sendMsgSubmitQueryResponse({ rootGetters }, { value, fee = [], memo = '' }) {
			try {
				const txClient=await initTxClient(rootGetters)
				const msg = await txClient.msgSubmitQueryResponse(value)
				const result = await txClient.signAndBroadcast([msg], {fee: { amount: fee, 
	gas: "200000" }, memo})
				return result
			} catch (e) {
				if (e == MissingWalletError) {
					throw new Error('TxClient:MsgSubmitQueryResponse:Init Could not initialize signing client. Wallet is required.')
				}else{
					throw new Error('TxClient:MsgSubmitQueryResponse:Send Could not broadcast Tx: '+ e.message)
				}
			}
		},
		
		async MsgQueryBalance({ rootGetters }, { value }) {
			try {
				const txClient=await initTxClient(rootGetters)
				const msg = await txClient.msgQueryBalance(value)
				return msg
			} catch (e) {
				if (e == MissingWalletError) {
					throw new Error('TxClient:MsgQueryBalance:Init Could not initialize signing client. Wallet is required.')
				} else{
					throw new Error('TxClient:MsgQueryBalance:Create Could not create message: ' + e.message)
				}
			}
		},
		async MsgSubmitTx({ rootGetters }, { value }) {
			try {
				const txClient=await initTxClient(rootGetters)
				const msg = await txClient.msgSubmitTx(value)
				return msg
			} catch (e) {
				if (e == MissingWalletError) {
					throw new Error('TxClient:MsgSubmitTx:Init Could not initialize signing client. Wallet is required.')
				} else{
					throw new Error('TxClient:MsgSubmitTx:Create Could not create message: ' + e.message)
				}
			}
		},
		async MsgSubmitQueryResponse({ rootGetters }, { value }) {
			try {
				const txClient=await initTxClient(rootGetters)
				const msg = await txClient.msgSubmitQueryResponse(value)
				return msg
			} catch (e) {
				if (e == MissingWalletError) {
					throw new Error('TxClient:MsgSubmitQueryResponse:Init Could not initialize signing client. Wallet is required.')
				} else{
					throw new Error('TxClient:MsgSubmitQueryResponse:Create Could not create message: ' + e.message)
				}
			}
		},
		
	}
}
