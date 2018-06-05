import React, { Component } from 'react'
import BlocksView from './BlocksView.jsx'

// gets the list of accounts out of the raw block data
export default class BlocksController extends Component {
    render() {
        const {
            allBlocks,
            filteredBlocks
        } = this.props

        return (
            <BlocksView allBlocks={allBlocks} filteredBlocks={filteredBlocks} />
        )
    }
}