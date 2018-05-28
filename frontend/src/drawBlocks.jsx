import React from 'react'
import { Card, CardBody, CardTitle } from 'reactstrap'

export default function DrawBlocks(blockData) {
    let blocks = []
    blockData.map((block) => {
        <Card>
        <CardTitle>Action: {block.action}</CardTitle>
        <CardTitle>Account: {block.account}</CardTitle>
        <CardTitle>Token: {block.token}</CardTitle>
        <CardTitle>Previous: {block.previous}</CardTitle>
        <CardTitle>Representative: {block.representative}</CardTitle>
        <CardTitle>Balance: {block.balance}</CardTitle>
        <CardTitle>Link: {block.link}</CardTitle>
        <CardTitle>Signature: {block.signature}</CardTitle>
        </Card>
    })
    return (
        {blocks}
    )
}