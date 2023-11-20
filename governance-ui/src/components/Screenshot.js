import React, { useRef, useState } from "react";
import html2canvas from "html2canvas";
import MainPage from "./MainPage";
import { Button, TextField } from "@mui/material";
import { DatePicker } from "@mui/x-date-pickers/DatePicker";
import { AdapterDayjs } from "@mui/x-date-pickers/AdapterDayjs";
import { LocalizationProvider } from "@mui/x-date-pickers/LocalizationProvider";
import { Grid, Box } from "@mui/material";
import dayjs from "dayjs";

const MyWebpage = () => {
  const contentRef = useRef(null);
  const [from, setFrom] = useState(dayjs(Date.now()));
  const [to, setTo] = useState(dayjs(Date.now()));

  const handleScreenshot = () => {
    if (contentRef.current) {
        html2canvas(contentRef.current, { scale: 2, allowTaint: true, useCORS: true }).then((canvas) => {
          // Convert the canvas to an image URL
          const screenshotUrl = canvas.toDataURL("image/png");
          // Create a link element to download the image
          const downloadLink = document.createElement("a");
          downloadLink.href = screenshotUrl;
          downloadLink.download = "screenshot.png";
          downloadLink.click();
        });
    }
  };

  const handleFrom = (date) => setFrom(date);

  const handleTo = (date) => setTo(date);

  return (
    <>
      <Grid
        container
        spacing={2}
        sx={{
          pt: 2,
          background: "#efefef",
        }}
      >
        <Grid item md={6} xs={12}></Grid>
        <Grid item xs={4} md={2} style={{ textAlign: "right" }}>
          <LocalizationProvider dateAdapter={AdapterDayjs}>
            <DatePicker
              label="From"
              value={from}
              disableFuture
              format="DD/MM/YYYY"
              onChange={handleFrom}
              slotProps={{ textField: { size: "small" } }}
              renderInput={(params) => <TextField {...params} />}
            />
          </LocalizationProvider>
        </Grid>

        <Grid item xs={4} md={2}>
          <LocalizationProvider dateAdapter={AdapterDayjs}>
            <Box textAlign="left">
              <DatePicker
                label="To"
                format="DD/MM/YYYY"
                disableFuture
                value={to}
                onChange={handleTo}
                slotProps={{ textField: { size: "small" } }}
                renderInput={(params) => <TextField {...params} />}
              />
            </Box>
          </LocalizationProvider>
        </Grid>
        <Grid item md={2} xs={4}>
          <Button
            onClick={handleScreenshot}
            variant="contained"
            disableElevation
            sx={{
              mt: 0.5,
            }}
            size="small"
          >
            Download report
          </Button>
        </Grid>
      </Grid>
      <div ref={contentRef}>
        <MainPage from={from} to={to} />
      </div>
    </>
  );
};

export default MyWebpage;
