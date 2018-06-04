import React from 'react'
import PropTypes from 'prop-types'
import { withStyles } from '@material-ui/core/styles'
import Typography from '@material-ui/core/Typography'
import Card from '@material-ui/core/Card'
import CardContent from '@material-ui/core/CardContent'
import Grid from '@material-ui/core/Grid'
import IconButton from '@material-ui/core/IconButton'
import ContentCopyIcon from '@material-ui/icons/ContentCopy'

import BlockDiagram from './BlockDiagram'

const styles = {
    accountHeader: {
        fontSize: 16
    },
    cardDiv: {
        display: "block",
        marginLeft: "15px",
        marginRight: "auto",
        textAlign: "center",
        marginTop: "15px",
        wordBreak: "all",
    },
}

function BlocksView(props) {
    const {
        activeBlock,
        blocks,
        classes,
        handleClick,
    } = props

    if (blocks === undefined) {
        return <div>No Blocks</div>
    }

    const blockList = blocks.map((block) => {
        const {
            Account,
            Action,
            Balance,
            Hash,
            Link,
            Previous,
            Representative,
            Signature,
            Token
        } = block

        function shortenText(display, text) {
            // base case
            if (text === "" || text === undefined) {
                return (
                    <div key={display}>
                        <Typography component="p">
                            <b>{display}:</b> null
                        </Typography>
                    </div>
                )
            }
            //shorten if too long
            const shortText = text.length > 15 ? text.substring(0, 15) + "..." : text
            return (
                <div key={display}>
                    <Typography component="p">
                        <b>{display}:</b> {shortText}
                    </Typography>
                </div>
            )
        }

        const cardValues = [
            ["Account", Account],
            ["Action", Action],
            ["Balance", Balance],
            ["Hash", Hash],
            ["Link", Link],
            ["Previous", Previous],
            ["Representative", Representative],
            ["Signature", Signature],
            ["Token", Token],
        ]
        const cardContent = cardValues.map((tuple) =>{
            return shortenText(tuple[0], tuple[1])
        })

        return (
            <div key={Hash} className={classes.cardDiv} onClick={() => handleClick(block)}>
                <Card className={classes.accountCard}>
                    <CardContent>
                        {cardContent}
                    </CardContent>
                </Card>
            </div>
        )
    })

    return (
        <div className={classes.grid}>
            <Grid container spacing={24}>
                <Grid item xs={3}>
                    {blockList}
                </Grid>
                <Grid item xs={9}>
                    <BlockDiagram block={activeBlock} />
                </Grid>
            </Grid>
        </div>
    )
}

BlocksView.propTypes = {
    classes: PropTypes.object.isRequired,
  }
  
  export default withStyles(styles)(BlocksView)