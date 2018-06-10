import React from 'react'
import {
	DiagramEngine,
	DiagramModel,
	DefaultNodeModel,
	DiagramWidget,
} from "storm-react-diagrams"

export default function BlockDiagram(props) {
    const {
        accounts
    } = props
    // setup
	var engine = new DiagramEngine()
	engine.installDefaultFactories()
	var model = new DiagramModel()

	var node1 = new DefaultNodeModel("Node 1", "rgb(0,192,255)");
	let port1 = node1.addOutPort("Out");
	node1.setPosition(100, 100);

	var node2 = new DefaultNodeModel("Node 2", "rgb(192,255,0)");
	let port2 = node2.addInPort("In");
	node2.setPosition(400, 100);

	// link the ports
	let link1 = port1.link(port2);
	link1.addLabel("Hello World!");

	//4) add the models to the root graph
	// model.addAll(node1, node2, link1);

    if (accounts) {
        var y = 25
        var nodes = []
        Object.keys(accounts).forEach((key) => {
            const blocks = accounts[key]
            const newNodes = makeNodes(blocks, y)
            nodes = nodes.concat(newNodes)
            y += 100
        })
        makePorts(nodes) 
        const links = makeLinks(nodes)
        
        nodes.forEach((node) => {
            model.addAll(node)
        })
        links.forEach((link) => {
            model.addAll(link)
        })
        
    }

	// 7) load model into engine
	engine.setDiagramModel(model);
	return <DiagramWidget className={"srd-demo-canvas"} diagramEngine={engine}  />
}

function makeNodes(blocks, y) {
    var x = 25
    return blocks.map((block) => {
        const shortAccount = shortenText(block.Account, true)
        var node = new DefaultNodeModel(block.Action + " by " + shortAccount, "rgb(0,192,255)")
        node.setPosition(x, y)
        node.block = block
        x = x + 250
        return node
    })
}

function makePorts(nodes) {
    const ports = []
    nodes.forEach((node) => {
        const action = node.block.Action
        if (node.block.Previous) {
            let portPrev = node.addInPort("Prev-----")
            ports.push(portPrev)
        } else {
            let portDefault = node.addInPort(" ")
            ports.push(portDefault)
        }
        if (action === "send") {
            const shortText = shortenText(node.block.Link, true)
            const portTo = node.addOutPort("To: " + shortText)
            ports.push(portTo)
        } else if (action === "open" || action === "receive") {
            const shortText = shortenText(node.block.Link, false)
            const portFrom = node.addOutPort("From: " + shortText)
            ports.push(portFrom)
        } else if (action === "issue") {
            const shortText = shortenText(node.block.Token, false)
            const portToken = node.addOutPort("Token: " + shortText)
            ports.push(portToken)
        } else if (action === "create-order") {
            const shortText = shortenText(node.block.Link, false)
            const portToken = node.addOutPort("Link: " + shortText)
            ports.push(portToken)
        } else if (action === "offer") {
            const shortText = shortenText(node.block.Left, false)
            const portToken = node.addOutPort("Left: " + shortText)
            ports.push(portToken)
        }

        let portLink = node.addOutPort("Link")
        ports.push(portLink)
    })
    return ports
}

function makeLinks(nodes) {
    const links = []
    nodes.forEach((node) => {
        const block = node.block
        var toNode = null
        if (block.Previous) {
            const prevPort = node.getOutPorts()[0]
            toNode = getHashNode(nodes, block.Previous)
            if (toNode) {
                const toPort = toNode.getInPorts()[0]
                const link = prevPort.link(toPort)
                links.push(link)
            }
        }
        if (block.Link) {
            if (block.Action === "send") {
                return
            } else if (block.Action === "open" || block.Action === "receive") {
                const prevPort = node.getOutPorts()[0]
                toNode = getHashNode(nodes, block.Link)
                if (toNode) {
                    const toPort = toNode.getInPorts()[0]
                    const link = prevPort.link(toPort)
                    links.push(link)
                }
            }
        }
        if (block.Left && block.Action === "offer") {
            const prevPort = node.getOutPorts()[0]
            toNode = getHashNode(nodes, block.Left)
            if (toNode) {
                const toPort = toNode.getInPorts()[0]
                const link = prevPort.link(toPort)
                links.push(link)
            }
        }
        if (block.Right && block.Action === "commit") {
            const prevPort = node.getOutPorts()[0]
            toNode = getHashNode(nodes, block.Right)
            if (toNode) {
                const toPort = toNode.getInPorts()[0]
                const link = prevPort.link(toPort)
                links.push(link)
            }
        }
        if (block.Link && block.Action === "create-order") {
            const prevPort = node.getOutPorts()[0]
            toNode = getHashNode(nodes, block.Link)
            if (toNode) {
                const toPort = toNode.getInPorts()[0]
                const link = prevPort.link(toPort)
                links.push(link)
            }
        }
        if (block.Link && block.Action === "offer") {
            const prevPort = node.getOutPorts()[0]
            toNode = getHashNode(nodes, block.Left)
            if (toNode) {
                const toPort = toNode.getInPorts()[0]
                const link = prevPort.link(toPort)
                links.push(link)
            }
        }
    })
    return links
}

function getHashNode(nodes, hash) {
    let returnNode = null
    nodes.forEach((node) => {
        if (node.block.Hash === hash) {
            returnNode = node
        }
    })
    return returnNode
}

function shortenText(text, address) {
	// base case
	if (text === "" || text === undefined) {
		return (
			null
		)
    }
    let withMods = text
    if (address) {
        withMods = text.substring(4, text.length)
    }
	//shorten if too long
	const isTooLong = withMods.length > 15
	const shortText = isTooLong ? withMods.substring(0, 15) + "..." : withMods
	return shortText
}