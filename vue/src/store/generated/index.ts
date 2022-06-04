// THIS FILE IS GENERATED AUTOMATICALLY. DO NOT MODIFY.

import StrideLabsStrideStridelabsStrideEpochs from './Stride-Labs/stride/Stridelabs.stride.epochs'
import StrideLabsStrideStridelabsStrideStakeibc from './Stride-Labs/stride/Stridelabs.stride.stakeibc'
import StrideLabsStrideStrideInterchainquery from './Stride-Labs/stride/stride.interchainquery'
import StrideLabsStrideStrideMintV1Beta1 from './Stride-Labs/stride/stride.mint.v1beta1'


export default { 
  StrideLabsStrideStridelabsStrideEpochs: load(StrideLabsStrideStridelabsStrideEpochs, 'Stridelabs.stride.epochs'),
  StrideLabsStrideStridelabsStrideStakeibc: load(StrideLabsStrideStridelabsStrideStakeibc, 'Stridelabs.stride.stakeibc'),
  StrideLabsStrideStrideInterchainquery: load(StrideLabsStrideStrideInterchainquery, 'stride.interchainquery'),
  StrideLabsStrideStrideMintV1Beta1: load(StrideLabsStrideStrideMintV1Beta1, 'stride.mint.v1beta1'),
  
}


function load(mod, fullns) {
    return function init(store) {        
        if (store.hasModule([fullns])) {
            throw new Error('Duplicate module name detected: '+ fullns)
        }else{
            store.registerModule([fullns], mod)
            store.subscribe((mutation) => {
                if (mutation.type == 'common/env/INITIALIZE_WS_COMPLETE') {
                    store.dispatch(fullns+ '/init', null, {
                        root: true
                    })
                }
            })
        }
    }
}
