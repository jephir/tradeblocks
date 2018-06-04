import React, { Component } from 'react'
import BlocksView from './BlocksView.jsx'

// gets the list of accounts out of the raw block data
export default class BlocksController extends Component {
    state = {
        activeBlock: {},
    }

    componentDidMount() {
        const {
            blocks
        } = this.props
        if (blocks === undefined || blocks.length <= 0) {
            this.setState({
                activeBlock: undefined
            })
        } else {
            this.setState({
                activeBlock: blocks[0]
            })
        }
    }

    componentDidUpdate(prevProps, prevState) {
        if (prevProps.blocks != this.props.blocks) {
            // changing blocks
            this.setState({
                activeBlock: {}
            })
        }
    }

    render() {
        const {
            activeBlock,
        } = this.state
        const {
            blocks,
        } = this.props

        return (
            <BlocksView blocks={blocks} activeBlock={activeBlock}
                        handleClick={this.handleClick.bind(this)}/>
        )
    }

    handleClick(block) {
        this.setState({
            activeBlock: block
        })
    }
}