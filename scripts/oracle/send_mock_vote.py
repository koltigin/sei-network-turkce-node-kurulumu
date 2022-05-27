import subprocess
import requests
import sys
import time

prevote_period_to_be_committed = -1
last_committed_vote_period = -1

GET_VOTE_PERIOD_COMMAND = "./build/seid query params subspace oracle VotePeriod"
SEND_PREVOTE_COMMAND_TMPLT = ("./build/seid tx oracle aggregate-prevote abc 100uusdc,50uatom "
"{val_address} -y --from={key_name} --chain-id=sei-chain "
"--fees=10000000usei --gas=500000 --broadcast-mode=block")
SEND_VOTE_COMMAND_TMPLT = ("./build/seid tx oracle aggregate-vote abc 100uusdc,50uatom "
"{val_address} -y --from={key_name} --chain-id=sei-chain "
"--fees=10000000usei --gas=500000 --broadcast-mode=block")

def get_vote_period_param():
    vote_period_output = subprocess.check_output(
        [GET_VOTE_PERIOD_COMMAND],
        stderr=subprocess.STDOUT,
        shell=True,
    ).decode()
    return int(
        vote_period_output.split('\n')[2].split(' ')[1][2:-2]
    )

def get_current_block_height():
    res = requests.get(f"http://localhost:26657/block")
    block = res.json()["result"]["block"]
    return int(block["header"]["height"])

def get_vote_period():
    return get_current_block_height() // get_vote_period_param()

def is_output_successful(output):
    return output.split('\n')[0].split(' ')[1] == '0'

def get_height(output):
    return [int(l.split(" ")[1][1:-1]) for l in output.split('\n') if l.startswith("height")][0]

def send_vote(validator_addr, key_name, current_vote_period):
    global last_committed_vote_period

    vote_output = subprocess.check_output(
        [SEND_VOTE_COMMAND_TMPLT.format(val_address=validator_addr, key_name=key_name)],
        stderr=subprocess.STDOUT,
        shell=True,
    ).decode()
    if not is_output_successful(vote_output):
        print(vote_output)
        exit()
    last_committed_vote_period = current_vote_period

def send_prevote(validator_addr, key_name):
    global prevote_period_to_be_committed

    vote_output = subprocess.check_output(
        [SEND_PREVOTE_COMMAND_TMPLT.format(val_address=validator_addr, key_name=key_name)],
        stderr=subprocess.STDOUT,
        shell=True,
    ).decode()
    if not is_output_successful(vote_output):
        print(vote_output)
        exit()
    prevote_period_to_be_committed = get_height(vote_output) // get_vote_period_param()

def send(validator_addr, key_name):
    current_vote_period = get_vote_period()
    if current_vote_period > prevote_period_to_be_committed and prevote_period_to_be_committed != -1:
        print(f"Committing vote for period {prevote_period_to_be_committed}")
        send_vote(validator_addr, key_name, prevote_period_to_be_committed)
    if prevote_period_to_be_committed == last_committed_vote_period and current_vote_period > prevote_period_to_be_committed:
        print(f"Pre-voting for period {current_vote_period}")
        send_prevote(validator_addr, key_name)


def main():
    args = sys.argv[1:]
    validator_addr = args[0] # this is usually what `./build/seid query staking validators | grep operator_address` returns
    key_name = args[1] # the genesis key name (something like `admin` or `tony`)
    if len(args) == 2:
        interval = 10
    else:
        interval = int(args[2])
    while True:
        send(validator_addr, key_name)
        time.sleep(interval)

if __name__ == "__main__":
    main()
