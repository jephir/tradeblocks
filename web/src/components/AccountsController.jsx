import React, { Component } from 'react'
import AccountsView from './AccountsView.jsx'

// gets the list of accounts out of the raw block data
export default class AccountsController extends Component {
    state = {
        accounts: {},
        activeAccount: ""
    }

    componentDidMount() {
        const accountsDict = this.getAccounts(this.props.blocks)
        const firstAccount = Object.keys(accountsDict)[0]
        const demoBlocks = this.getDemoBlocks()
        this.setState({
            accounts: accountsDict,
            activeAccount: firstAccount,
            demoBlocks: demoBlocks,
        })
    }

    componentDidUpdate(prevProps, prevState) {
		if (prevProps.blocks !== this.props.blocks) {
			const accountsDict = this.getAccounts(this.props.blocks)
            const firstAccount = Object.keys(accountsDict)[0]
            const demoBlocks = this.getDemoBlocks()
            this.setState({
                accounts: accountsDict,
                activeAccount: firstAccount,
                demoBlocks: demoBlocks,
            })
		}
    }

    render() {
        return (
            <AccountsView accounts={this.state.accounts} allBlocks={this.props.blocks}
                          activeAccount={this.state.activeAccount}
                          handleClick={this.handleClick.bind(this)}/>
        )
    }

    getAccounts(blocks) {
        const accountsDict = {}
        this.props.blocks.forEach(function(block) {
            const blockList = accountsDict[block.Account] || []
            blockList.push(block)
            accountsDict[block.Account] = blockList
        })
        return accountsDict
    }

    handleClick(newAccount) {
        this.setState({
            activeAccount: newAccount
        })
    }

    getDemoBlocks() {
        const demoBlocks = []

        if (demoBlocks) {
            // issue for account1
            demoBlocks.push({
                Account: "xtb:Account1",
                Action: "issue",
                Balance: 1000,
                Hash: "Issue1_Hash",
                Link: "",
                Previous: "",
                Signature: "Issue1_Signature",
                Token: "Demo_Token1",
            })
    
            // send to account 2
            demoBlocks.push({
                Account: "xtb:Account1",
                Action: "send",
                Balance: 1000,
                Hash: "Send1_Hash",
                Link: "xtb:Account2",
                Previous: "Issue1_Hash",
                Signature: "Send1_Signature",
                Token: "Demo_Token1",
            })
    
            // make an open
            demoBlocks.push({
                Account: "xtb:Account2",
                Action: "open",
                Balance: 1000,
                Hash: "Open1_Hash",
                Link: "Send1_Hash",
                Previous: "",
                Signature: "Open1_Signature",
                Token: "Demo_Token1",
            })

            // make an send back to 1
            demoBlocks.push({
                Account: "xtb:Account2",
                Action: "send",
                Balance: 900,
                Hash: "Send2_Hash",
                Link: "xtb:Account1",
                Previous: "Open1_Hash",
                Signature: "Send2_Signature",
                Token: "Demo_Token1",
            })

            // receive the send
            demoBlocks.push({
                Account: "xtb:Account1",
                Action: "receive",
                Balance: 1100,
                Hash: "Receive1_Hash",
                Link: "Send2_Hash",
                Previous: "Send1_Hash",
                Signature: "Receive1_Signature",
                Token: "Demo_Token1",
            })

            // send for swap offer
            demoBlocks.push({
                Account: "xtb:Account1",
                Action: "send",
                Balance: 1100,
                Hash: "Send3_Hash",
                Link: "xtb:Swap1",
                Previous: "Receive1_Hash",
                Signature: "Receive1_Signature",
                Token: "Demo_Token1",
            })

            // swap offer
            demoBlocks.push({
                Account: "xtb:Swap1",
                Action: "offer",
                Balance: 1100,
                Hash: "Swap1_Hash",
                Left: "Send3_Hash",
                Previous: "",
                Signature: "Swap1_Signature",
                Token: "Demo_Token1",
            })

            // send for swap offer
            demoBlocks.push({
                Account: "xtb:Account2",
                Action: "send",
                Balance: 1100,
                Hash: "Send4_Hash",
                Link: "xtb:Swap1",
                Previous: "Send2_Hash",
                Signature: "Receive1_Signature",
                Token: "Demo_Token1",
            })

            // swap commit
            demoBlocks.push({
                Account: "xtb:Swap1",
                Action: "commit",
                Balance: 1100,
                Hash: "Swap2_Hash",
                Left: "Send3_Hash",
                Right: "Send4_Hash",
                Previous: "Swap1_Hash",
                Signature: "Swap2_Signature",
                Token: "Demo_Token1",
            })

            // receive from commit Account1
            demoBlocks.push({
                Account: "xtb:Account1",
                Action: "receive",
                Balance: 1100,
                Hash: "Receive2_Hash",
                Link: "Swap2_Hash",
                Previous: "Send3_Hash",
                Signature: "Receive2_Signature",
                Token: "Demo_Token1",
            })

            // receive from commit Account2
            demoBlocks.push({
                Account: "xtb:Account2",
                Action: "receive",
                Balance: 1100,
                Hash: "Receive3_Hash",
                Link: "Swap2_Hash",
                Previous: "Send4_Hash",
                Signature: "Receive2_Signature",
                Token: "Demo_Token1",
            })
        }
        return demoBlocks
    }
}