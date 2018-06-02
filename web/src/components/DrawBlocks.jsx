import React from 'react'
import ReactJson from 'react-json-view'

export default function DrawBlocks(props) {
    const { 
        blocks,
    } = props
    if (blocks === undefined) {
        return <div>Loading...</div>
    }

    const blockText = blocks.map((block) => {
        return (
            <div key={block.Hash}>
                <ReactJson src={block} />
            </div>
        )
    })
    return (
        <div>
            {blockText}
        </div>
    )
}