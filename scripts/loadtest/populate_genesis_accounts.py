import json
import os
import subprocess
import sys

def add_genesis_account(account_name, amount, local=False):
    if local:
        add_key_cmd = f"yes | ./build/seid keys add {account_name}"
    else:
        add_key_cmd = f"printf '12345678\n' | ./build/seid keys add {account_name}"
    add_key_output = subprocess.check_output(
        [add_key_cmd],
        stderr=subprocess.STDOUT,
        shell=True,
    ).decode()
    splitted_outputs = add_key_output.split('\n')
    address = splitted_outputs[3].split(': ')[1]
    mnemonic = splitted_outputs[11]
    if local:
        add_account_cmd = f"./build/seid add-genesis-account {address} {amount}"
    else:
        add_account_cmd = f"printf '12345678\n' | ./build/seid add-genesis-account {address} {amount}"

    home_path = os.path.expanduser('~')
    filename = f"{home_path}/test_accounts/{account_name}.json"
    os.makedirs(os.path.dirname(filename), exist_ok=True)
    with open(filename, 'w') as f:
        data = {
            "address": address,
            "mnemonic": mnemonic,
        }
        json.dump(data, f)
    subprocess.check_call(
        [add_account_cmd],
        shell=True,
    )

def bulk_create_genesis_accounts(number_of_accounts, amount, is_local=False):
    for i in range(number_of_accounts):
        print(f"Creating account {i}")
        add_genesis_account(f"ta{i}", amount, is_local)

def main():
    args = sys.argv[1:]
    if args[0] == "single":
        add_genesis_account(args[1], args[2], True)
        return
    number_of_accounts = int(args[0])
    is_local = False
    if len(args) > 2 and args[2] == "loc":
        is_local = True
    bulk_create_genesis_accounts(number_of_accounts, args[1], is_local)

if __name__ == "__main__":
    main()
