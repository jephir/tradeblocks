import React, { Componenet } from 'react'

export default class App extends Component {
    state = {
        blocks: []
    }

    componentDidMount() {
        this.getBlocks()
    }

    getBlocks() {
        $.ajax("/blocks", {
            method: "get",
            dataType: "json",
            contentType: "application/json; charset=UTF-8",
            success: (resp) => {
                this.setState({
                    blocks: resp.data
                })
            }
        })
    }

    render() {
        return <DrawBlocks blocks={this.state.blocks} />
    }

}

