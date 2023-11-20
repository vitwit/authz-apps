import { Grid, Typography } from "@mui/material";

function Text({ proposalId, title, voteOption }) {
  return (
    <Grid container alignItems="center" spacing={2}>
      <Grid item xs={10}>
        <Typography
          style={{
            wordWrap: "break-word",
          }}
          variant="body1"
          fontWeight={500}
        >
          <b>#{proposalId}</b>&nbsp;&nbsp;{title}
        </Typography>
      </Grid>
      <Grid item xs={2}>
        <Typography
          variant="body2"
          style={{
            color: getColor(voteOption),
            wordWrap: "break-word",
          }}
          fontWeight={600}
        >
          {parseVoteOption(voteOption)}
        </Typography>
      </Grid>
    </Grid>
  );
}

export default Text;

const parseVoteOption = (option) => {
  switch (option) {
    case "VOTE_OPTION_YES":
      return "YES";
    case "VOTE_OPTION_NO":
      return "NO";
    case "VOTE_OPTION_ABSTAIN":
      return "ABSTAIN";
    case "VOTE_OPTION_NO_WITH_VETO":
      return "VETO";
    default:
      return "-";
  }
};

const getColor = (option) => {
  switch (option) {
    case "VOTE_OPTION_YES":
      return "green";
    case "VOTE_OPTION_NO":
      return "indianred";
    case "VOTE_OPTION_ABSTAIN":
      return "gray";
    case "VOTE_OPTION_NO_WITH_VETO":
      return "red";
    default:
      return "gray";
  }
};
