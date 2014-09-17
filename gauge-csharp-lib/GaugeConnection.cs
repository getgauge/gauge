using System;
using System.Collections.Generic;
using System.Net;
using System.Net.Sockets;
using Google.ProtocolBuffers;
using main;

namespace gauge_csharp_lib
{
    public class GaugeConnection
    {
        private readonly int _port;
        private readonly TcpClient _tcpCilent;

        public GaugeConnection(int port)
        {
            var tcpClient = new TcpClient();
            try
            {
                tcpClient.Connect(new IPEndPoint(IPAddress.Loopback, port));
            }
            catch (Exception e)
            {
                throw new Exception("Could not connect", e);
            }
            _tcpCilent = tcpClient;
            _port = port;
        }

        public int Port
        {
            get { return _port; }
        }

        public IList<ProtoStepValue> fetchAllSteps()
        {
            GetAllStepsRequest getAllStepsRequest = GetAllStepsRequest.CreateBuilder().Build();
            APIMessage fetchStepsApiRequest =
                APIMessage.CreateBuilder()
                    .SetAllStepsRequest(getAllStepsRequest)
                    .SetMessageId(1)
                    .SetMessageType(APIMessage.Types.APIMessageType.GetAllStepsRequest)
                    .Build();
            NetworkStream networkStream = _tcpCilent.GetStream();

            APIMessage apiMessage = WriteAndReadApiMessage(fetchStepsApiRequest, networkStream);
            return apiMessage.AllStepsResponse.AllStepsList;
        }


        private static APIMessage readMessage(NetworkStream networkStream)
        {
            byte[] responseBytes = readBytes(networkStream);
            APIMessage apiMessage = APIMessage.ParseFrom(responseBytes);
            return apiMessage;
        }

        private static void writeAPIMessage(APIMessage apiRequest, NetworkStream networkStream)
        {
            byte[] bytes = apiRequest.ToByteArray();
            CodedOutputStream cos = CodedOutputStream.CreateInstance(networkStream);
            cos.WriteRawVarint64((ulong) bytes.Length);
            cos.Flush();
            networkStream.Write(bytes, 0, bytes.Length);
            networkStream.Flush();
        }

        private static byte[] readBytes(NetworkStream networkStream)
        {
            CodedInputStream codedInputStream = CodedInputStream.CreateInstance(networkStream);
            ulong messageLength = codedInputStream.ReadRawVarint64();
            var bytes = new List<byte>();
            for (ulong i = 0; i < messageLength; i++)
            {
                bytes.Add(codedInputStream.ReadRawByte());
            }
            return bytes.ToArray();
        }

        public string getStepValue(string stepText, bool hasInlineTable)
        {
            GetStepValueRequest stepValueRequest =
                GetStepValueRequest.CreateBuilder().SetStepText(stepText).SetHasInlineTable(hasInlineTable).Build();
            APIMessage stepValueRequestMessage =
                APIMessage.CreateBuilder()
                    .SetMessageId(2)
                    .SetMessageType(APIMessage.Types.APIMessageType.GetStepValueRequest)
                    .SetStepValueRequest(stepValueRequest)
                    .Build();
            NetworkStream networkStream = _tcpCilent.GetStream();
            APIMessage apiMessage = WriteAndReadApiMessage(stepValueRequestMessage, networkStream);
            return apiMessage.StepValueResponse.StepValue.StepValue;
        }

        private APIMessage WriteAndReadApiMessage(APIMessage stepValueRequestMessage, NetworkStream networkStream)
        {
            lock (_tcpCilent)
            {
                writeAPIMessage(stepValueRequestMessage, networkStream);
                APIMessage apiMessage = readMessage(networkStream);
                return apiMessage;
            }
        }
    }
}