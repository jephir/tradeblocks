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
        this.setState({
            accounts: accountsDict,
            activeAccount: firstAccount
        })
    }

    componentDidUpdate(prevProps, prevState) {
		if (prevProps.blocks !== this.props.blocks) {
			const accountsDict = this.getAccounts(this.props.blocks)
            const firstAccount = Object.keys(accountsDict)[0]
            this.setState({
                accounts: accountsDict,
                activeAccount: firstAccount
            })
		}
    }

    render() {
        return (
            <AccountsView accounts={this.state.accounts} 
                          activeAccount={this.state.activeAccount}
                          handleClick={this.handleClick.bind(this)}/>
        )
    }

    getAccounts(blocks) {
        const accountsDict = {}
        blocks.forEach(function(block) {
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
}