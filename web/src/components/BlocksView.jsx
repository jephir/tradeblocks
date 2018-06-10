import React from 'react'
import PropTypes from 'prop-types'
import { withStyles } from '@material-ui/core/styles'
import Typography from '@material-ui/core/Typography'
import Card from '@material-ui/core/Card'
import CardContent from '@material-ui/core/CardContent'
import ContentCopyIcon from '@material-ui/icons/ContentCopy'

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
    icon: {
        fontSize: 14,
    }
}

function BlocksView(props) {
    const {
        filteredBlocks,
        classes,
    } = props

    if (filteredBlocks === undefined) {
        return <div>No Blocks</div>
    }

    const blockList = filteredBlocks.map((block) => {
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
            const isTooLong = text.length > 17
            const shortText =  isTooLong ? text.substring(0, 15) + "..." : text
            return (
                <div key={display}>
                    <Typography component="p">
                        <b>{display}:</b> {shortText}
                        {isTooLong && <ContentCopyIcon className={classes.icon} /> }
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
            <div key={Hash} className={classes.cardDiv}>
                <Card className={classes.accountCard}>
                    <CardContent>
                        {cardContent}
                    </CardContent>
                </Card>
            </div>
        )
    })

    return (
        <div>
            {blockList}
        </div>
    )
}

BlocksView.propTypes = {
    classes: PropTypes.object.isRequired,
  }
  
  export default withStyles(styles)(BlocksView)