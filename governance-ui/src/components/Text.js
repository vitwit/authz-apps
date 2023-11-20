import { Grid, Typography } from "@mui/material";

function Text({ proposalId, title, voteOption }) {
  return (
    <Grid
      container
      alignItems="center"
      spacing={2}
      
    >
      <Grid item xs={2}>
        <Typography
          variant="body1"
          style={{
            wordWrap: "break-word",
          }}
        >
          #{proposalId}
        </Typography>
      </Grid>
      <Grid item xs={8}>
        <Typography
          style={{
            wordWrap: "break-word",
          }}
          variant="body1"
          fontWeight={500}
        >
          {title}
        </Typography>
      </Grid>
      <Grid item xs={2}>
        <Typography
          variant="body1"
          style={{
            color:
              voteOption === ("NO" || "NO_WITH_VETO")
                ? "red"
                : voteOption === "ABSTAIN"
                ? "gray"
                : "green",
            wordWrap: "break-word",
          }}
          fontWeight={600}
        >
          {voteOption}
        </Typography>
      </Grid>
    </Grid>
  );
}

export default Text;
