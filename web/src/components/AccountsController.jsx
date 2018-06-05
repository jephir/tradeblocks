import React, { Component } from 'react';
import DrawBlocks from './DrawBlocks.jsx'

// gets the list of accounts out of the raw block data
export default class AccountsController extends Component {
    state = {
        accounts: {}
    }

    componentDidMount() {
        const accountsDict = this.getAccounts(this.props.blocks)
        this.setState({
            accounts: accountsDict
        })
    }

    componentDidUpdate(prevProps, prevState) {
		if (prevProps.blocks !== this.props.blocks) {
			const accountsDict = this.getAccounts(this.props.blocks)
            this.setState({
                accounts: accountsDict
            })
		}
    }

    render() {
        return <DrawBlocks blocks={this.props.blocks} />
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
}