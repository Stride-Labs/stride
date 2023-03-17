import pandas as pd
import json

'''
    This file provides an extremely lightweight tool to compare two state exports from a "localstride" test.

    It will diff parts of state and will output any diffs that are not "expected"

    To run, please modify "BEFORE_PATH" and "AFTER_PATH" and run `python3 compare_before_after_state.py`
'''

BEFORE_PATH = 'state_export_before_upgrade.json'
AFTER_PATH = 'state_export_after_upgrade.json'

EXPECTED_DIFFS = [
    '/interchainquery/queries',
    '/auth/accounts',
    '/bank/balances',
    '/bank/supply',
    '/epochs/epochs',
    '/staking/last_total_power',
    '/staking/delegations',
    '/staking/validators',
    '/staking/last_validator_powers',
    '/staking/unbonding_delegations',
    '/gov/proposals',
    '/gov/voting_params/voting_period',
    '/gov/starting_proposal_id',
    '/interchainaccounts/controller_genesis_state/ports',
    '/slashing/missed_blocks',
    '/slashing/signing_infos',
    '/distribution/delegator_starting_infos',
    '/distribution/previous_proposer',
]

def compare_app_state(before_state, after_state):
    before = before_state['app_state']
    after = after_state['app_state']
    equal_paths, unequal_paths = _return_equal_unequal_paths_recursively(before, after, '')
    print("THE FOLLOWING STORES ARE UNEXPECTEDLY DIFFERENT - MUST INVESTIGATE")
    expected_diff = []
    for p in unequal_paths:
        if p in EXPECTED_DIFFS:
            expected_diff.append(p)
            continue
        print(f"\t{p}")
    print("\n\nTHE FOLLOWING STORES ARE DIFFERENT, AS EXPECTED - NO ACTION REQUIRED")
    for p in expected_diff:
        print(f"\t{p}")

def _clean_path(path_list):
    return [c.replace(':/', ':') for c in path_list]

def _return_equal_unequal_paths_recursively(s1, s2, prefix):
    if (type(s1) != dict) or (type(s2) != dict):
        if s1 == s2:
            return [prefix], []
        else:
            if (type(s1) == type(s2)) and (type(s1) == list) and (len(s1) == len(s2)) and (prefix not in EXPECTED_DIFFS):
                zipped_list = zip(s1, s2)
                equal_paths = set()
                unequal_paths = set()
                for zl in zipped_list:
                    sub_equal, sub_unequal = _return_equal_unequal_paths_recursively(zl[0], zl[1], prefix + ':')
                    equal_paths.update(sub_equal)
                    unequal_paths.update(sub_unequal)
                return _clean_path(list(equal_paths)), _clean_path(list(unequal_paths))
            return [], [prefix]
    equal_paths = []
    unequal_paths = []
    s1_keys = set(s1.keys())
    s2_keys = set(s2.keys())
    inter_keys = s1_keys.intersection(s2_keys)
    s1_unequal = s1_keys.difference(s2_keys)
    s2_unequal = s2_keys.difference(s1_keys)
    for joint_key in inter_keys:
        sub_equal, sub_unequal = _return_equal_unequal_paths_recursively(s1[joint_key], s2[joint_key], prefix + f'/{joint_key}')
        equal_paths.extend(sub_equal)
        unequal_paths.extend(sub_unequal)
    for s1_key in s1_unequal:
        unequal_paths.append(prefix + f'/{s1_key} - MISSING')
    for s2_key in s2_unequal:
        unequal_paths.append(prefix + f'/{s2_key} - ADDED')
 
    return equal_paths, unequal_paths

def main():
    with open(BEFORE_PATH) as f:
        before_state = json.load(f)
    with open(AFTER_PATH) as f:
        after_state = json.load(f)

    compare_app_state(before_state, after_state)

if __name__ == '__main__':  
    main()