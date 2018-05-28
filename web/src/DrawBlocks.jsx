import React from 'react'
import ReactJson from 'react-json-view'

export default function DrawBlocks(props) {
    const { 
        blocks,
    } = props
    console.log("blcoks", blocks)
    if (blocks === undefined) {
        return <div>Loading...</div>
    }

    const blockText = blocks.map((block) => {
        return (
            <div>
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