import { Masonry } from "@mui/lab";
import CustomCard from "./CustomCard";
import { Box, Typography } from "@mui/material";
import { useEffect, useState } from "react";

function CardList({ from, to }) {
  const [data, setData] = useState({});
  const [errorMessage, setErrorMessage] = useState("");

  useEffect(() => {
    fetch(
      `${process.env.REACT_APP_API_URI}/votes?start=${Math.floor(
        from.unix() / 1
      )}&end=${Math.floor(to.unix() / 1)}`,
      {
        method: "GET",
        headers: {
          "Content-Type": "application/json",
        },
      }
    )
      .then((response) => response.json())
      .then((data) => {
        setData(data);
      })
      .catch((error) => {
        setErrorMessage(error?.message);
      });
  }, [from, to]);

  return (
    <>
      <Box
        sx={{
          textAlign: "center",
          margin: "auto",
        }}
      >
        {errorMessage.length > 0 ? (
          <Typography variant="h5" fontWeight={600} color="red">
            {errorMessage}
          </Typography>
        ) : null}
      </Box>

      {Object.keys(data).length === 0 && errorMessage.length === 0 ? (
        <Typography
          variant="h5"
          fontWeight={500}
          textAlign="center"
          sx={{
            mt: 2,
          }}
        >
          Nothing here
        </Typography>
      ) : (
        <Masonry columns={2} spacing={2}>
          {Object.keys(data).map((name, index) => (
            <CustomCard key={index} networkName={name} data={data[name]} />
          ))}
        </Masonry>
      )}
    </>
  );
}

export default CardList;
